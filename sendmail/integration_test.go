//go:build integration
// +build integration

package sendmail

import (
	"encoding/base64"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

// getTestSMTPHost returns SMTP host from environment or default to localhost
func getTestSMTPHost() string {
	if host := os.Getenv("TEST_SMTP_HOST"); host != "" {
		return host
	}
	return "localhost"
}

// getTestSMTPPort returns SMTP port from environment or default to 1025
func getTestSMTPPort() string {
	if port := os.Getenv("TEST_SMTP_PORT"); port != "" {
		return port
	}
	return "1025"
}

// TestIntegrationSendSimpleMailToMailpit 測試發送簡單郵件到 Mailpit
func TestIntegrationSendSimpleMailToMailpit(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	viper.Reset()

	host := getTestSMTPHost()
	port := getTestSMTPPort()

	viper.Set("mailField", map[string]any{
		"host":     host,
		"from":     "test@example.com",
		"to":       "recipient@example.com",
		"cc":       "",
		"bcc":      "",
		"subject":  "Integration Test",
		"contents": "This is an integration test email",
		"port":     port,
	})

	ok, err := SendMailWithMultipart("mailField")
	if err != nil {
		t.Fatalf("SendMailWithMultipart failed: %v", err)
	}
	if !ok {
		t.Fatalf("SendMailWithMultipart returned false")
	}

	t.Log("✓ Email sent successfully to Mailpit")
}

// TestIntegrationSendComplexMailToMailpit 測試發送包含 To、Cc、Bcc 的複雜郵件
func TestIntegrationSendComplexMailToMailpit(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	viper.Reset()

	host := getTestSMTPHost()
	port := getTestSMTPPort()

	viper.Set("mailField", map[string]any{
		"host":     host,
		"from":     "sender@example.com",
		"to":       "recipient1@example.com,recipient2@example.com",
		"cc":       "cc@example.com",
		"bcc":      "bcc@example.com",
		"subject":  "複雜郵件測試",
		"contents": "This email has multiple recipients, CC, and BCC",
		"port":     port,
	})

	ok, err := SendMailWithMultipart("mailField")
	if err != nil {
		t.Fatalf("SendMailWithMultipart failed: %v", err)
	}
	if !ok {
		t.Fatalf("SendMailWithMultipart returned false")
	}

	t.Log("✓ Complex email sent successfully to Mailpit")
}

// TestIntegrationSendMailConnectsToCorrectHost 測試寄信連接到正確的主機和端口
func TestIntegrationSendMailConnectsToCorrectHost(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	host := getTestSMTPHost()
	port := getTestSMTPPort()

	var gotAddr string
	original := SendMail
	defer func() { SendMail = original }()

	SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		gotAddr = addr
		return nil
	}

	viper.Reset()
	viper.Set("mailField", map[string]any{
		"host":     host,
		"from":     "test@example.com",
		"to":       "recipient@example.com",
		"cc":       "",
		"bcc":      "",
		"subject":  "Connection test",
		"contents": "body",
		"port":     port,
	})

	ok, err := SendMailWithMultipart("mailField")
	if err != nil || !ok {
		t.Fatalf("SendMailWithMultipart failed: %v", err)
	}

	expectedAddr := host + ":" + port
	if gotAddr != expectedAddr {
		t.Errorf("Expected address %s, got %s", expectedAddr, gotAddr)
	}

	t.Logf("✓ Connected to correct host: %s", gotAddr)
}

// TestIntegrationEmailContentVerification 測試郵件內容驗證
func TestIntegrationEmailContentVerification(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	viper.Reset()

	host := getTestSMTPHost()
	port := getTestSMTPPort()
	testSubject := "Content Verification Test"
	testContent := "This is test content for verification"

	viper.Set("mailField", map[string]any{
		"host":     host,
		"from":     "sender@example.com",
		"to":       "recipient@example.com",
		"cc":       "",
		"bcc":      "",
		"subject":  testSubject,
		"contents": testContent,
		"port":     port,
	})

	ok, err := SendMailWithMultipart("mailField")
	if err != nil || !ok {
		t.Fatalf("SendMailWithMultipart failed: %v", err)
	}

	// 等待郵件被 Mailpit 索引
	time.Sleep(100 * time.Millisecond)

	// 驗證郵件內容
	message, err := getLatestMessage()
	if err != nil {
		t.Fatalf("Failed to get latest message: %v", err)
	}

	assertSubjectEquals(t, message, testSubject)
	assertFromEquals(t, message, "sender@example.com")
	assertToContains(t, message, []string{"recipient@example.com"})

	t.Log("✓ Email content verification passed")
}

// TestIntegrationComplexEmailContent 測試複雜郵件的內容驗證
func TestIntegrationComplexEmailContent(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	viper.Reset()

	host := getTestSMTPHost()
	port := getTestSMTPPort()

	viper.Set("mailField", map[string]any{
		"host":     host,
		"from":     "sender@example.com",
		"to":       "recipient1@example.com,recipient2@example.com",
		"cc":       "cc@example.com",
		"bcc":      "bcc@example.com",
		"subject":  "複雜郵件內容測試",
		"contents": "Complex email with multiple recipients",
		"port":     port,
	})

	ok, err := SendMailWithMultipart("mailField")
	if err != nil || !ok {
		t.Fatalf("SendMailWithMultipart failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	message, err := getLatestMessage()
	if err != nil {
		t.Fatalf("Failed to get latest message: %v", err)
	}

	// 驗證多個收件人
	assertToContains(t, message, []string{"recipient1@example.com", "recipient2@example.com"})
	assertCcContains(t, message, []string{"cc@example.com"})

	t.Log("✓ Complex email content verification passed")
}

// TestIntegrationChineseEncoding 測試中文主題編碼
func TestIntegrationChineseEncoding(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	viper.Reset()

	host := getTestSMTPHost()
	port := getTestSMTPPort()
	chineseSubject := "中文主題測試"

	viper.Set("mailField", map[string]any{
		"host":     host,
		"from":     "sender@example.com",
		"to":       "recipient@example.com",
		"cc":       "",
		"bcc":      "",
		"subject":  chineseSubject,
		"contents": "English content",
		"port":     port,
	})

	ok, err := SendMailWithMultipart("mailField")
	if err != nil || !ok {
		t.Fatalf("SendMailWithMultipart failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	message, err := getLatestMessage()
	if err != nil {
		t.Fatalf("Failed to get latest message: %v", err)
	}

	// 驗證中文主題被正確編碼和解碼
	assertSubjectEquals(t, message, chineseSubject)

	t.Log("✓ Chinese encoding verification passed")
}

// TestIntegrationChineseContent 測試中文正文內容
func TestIntegrationChineseContent(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	viper.Reset()

	host := getTestSMTPHost()
	port := getTestSMTPPort()
	chineseContent := "這是一個中文測試內容"

	viper.Set("mailField", map[string]any{
		"host":     host,
		"from":     "sender@example.com",
		"to":       "recipient@example.com",
		"cc":       "",
		"bcc":      "",
		"subject":  "Chinese Content Test",
		"contents": chineseContent,
		"port":     port,
	})

	ok, err := SendMailWithMultipart("mailField")
	if err != nil || !ok {
		t.Fatalf("SendMailWithMultipart failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	rawMessage, err := getRawMessage("latest")
	if err != nil {
		t.Fatalf("Failed to get raw message: %v", err)
	}

	// 驗證中文內容被正確編碼
	assertContentContains(t, rawMessage, chineseContent)

	t.Log("✓ Chinese content verification passed")
}

// TestIntegrationAttachmentInEmail 測試郵件附件驗證
func TestIntegrationAttachmentInEmail(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	// 創建臨時測試附件
	testAttachmentPath := t.TempDir() + "/test_attachment.txt"
	testAttachmentContent := "This is a test attachment content"

	err := os.WriteFile(testAttachmentPath, []byte(testAttachmentContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test attachment: %v", err)
	}

	viper.Reset()

	host := getTestSMTPHost()
	port := getTestSMTPPort()

	viper.Set("mailField", map[string]any{
		"host":       host,
		"from":       "sender@example.com",
		"to":         "recipient@example.com",
		"cc":         "",
		"bcc":        "",
		"subject":    "Attachment Test",
		"contents":   "Email with attachment",
		"port":       port,
		"attachment": testAttachmentPath,
	})

	ok, err := SendMailWithMultipart("mailField")
	if err != nil || !ok {
		t.Fatalf("SendMailWithMultipart failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	message, err := getLatestMessage()
	if err != nil {
		t.Fatalf("Failed to get latest message: %v", err)
	}

	// 驗證附件存在
	assertAttachmentExists(t, message, "test_attachment.txt")

	t.Log("✓ Attachment verification passed")
}

// TestIntegrationMIMEStructure 測試郵件的 MIME 結構
func TestIntegrationMIMEStructure(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	viper.Reset()

	host := getTestSMTPHost()
	port := getTestSMTPPort()

	viper.Set("mailField", map[string]any{
		"host":     host,
		"from":     "sender@example.com",
		"to":       "recipient@example.com",
		"cc":       "",
		"bcc":      "",
		"subject":  "MIME Structure Test",
		"contents": "Test content for MIME verification",
		"port":     port,
	})

	ok, err := SendMailWithMultipart("mailField")
	if err != nil || !ok {
		t.Fatalf("SendMailWithMultipart failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	rawMessage, err := getRawMessage("latest")
	if err != nil {
		t.Fatalf("Failed to get raw message: %v", err)
	}

	// 驗證 MIME 結構
	assertMIMEStructure(t, rawMessage, "multipart/mixed")

	t.Log("✓ MIME structure verification passed")
}

// TestIntegrationBurstModeSample 測試爆發模式發送少量郵件
func TestIntegrationBurstModeSample(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	host := getTestSMTPHost()
	port := getTestSMTPPort()
	quantity := 10 // 發送 10 封郵件

	// 獲取發送前的郵件計數
	initialCount, err := getMessageCount()
	if err != nil {
		t.Fatalf("Failed to get initial message count: %v", err)
	}

	// 執行爆發模式
	if err := BurstModeSendMail(BurstOptions{
		Quantity:       quantity,
		Host:           host,
		Port:           port,
		Domains:        []string{"example.com", "test.com"},
		AllowedDomains: []string{"example.com", "test.com"},
	}); err != nil {
		t.Fatalf("BurstModeSendMail failed: %v", err)
	}

	// 等待郵件被 Mailpit 索引
	time.Sleep(500 * time.Millisecond)

	// 獲取發送後的郵件計數
	finalCount, err := getMessageCount()
	if err != nil {
		t.Fatalf("Failed to get final message count: %v", err)
	}

	// 驗證郵件計數增加了
	newMessageCount := finalCount - initialCount
	if newMessageCount < quantity {
		t.Errorf("Expected at least %d messages, but got %d new messages", quantity, newMessageCount)
	}

	t.Logf("✓ Burst mode test passed: sent %d messages", newMessageCount)
}

// TestIntegrationCcWithInvalidAddress 測試 Cc 包含無效地址時拒絕發送
func TestIntegrationCcWithInvalidAddress(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	viper.Reset()

	host := getTestSMTPHost()
	port := getTestSMTPPort()

	viper.Set("mailField", map[string]any{
		"host":     host,
		"from":     "sender@example.com",
		"to":       "recipient@example.com",
		"cc":       "invalid@,valid@example.com",
		"bcc":      "",
		"subject":  "Cc Invalid Test",
		"contents": "This should not be sent",
		"port":     port,
	})

	ok, err := SendMailWithMultipart("mailField")
	if ok || err == nil {
		t.Fatalf("SendMailWithMultipart should return error for invalid cc, but didn't")
	}

	t.Logf("✓ Invalid Cc correctly rejected: %v", err)
}

// TestIntegrationBccWithInvalidAddress 測試 Bcc 包含無效地址時拒絕發送
func TestIntegrationBccWithInvalidAddress(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	viper.Reset()

	host := getTestSMTPHost()
	port := getTestSMTPPort()

	viper.Set("mailField", map[string]any{
		"host":     host,
		"from":     "sender@example.com",
		"to":       "recipient@example.com",
		"cc":       "",
		"bcc":      "bcc@invalid",
		"subject":  "Bcc Invalid Test",
		"contents": "This should not be sent",
		"port":     port,
	})

	ok, err := SendMailWithMultipart("mailField")
	if ok || err == nil {
		t.Fatalf("SendMailWithMultipart should return error for invalid bcc, but didn't")
	}

	t.Logf("✓ Invalid Bcc correctly rejected: %v", err)
}

// TestIntegrationMultipleRecipientsInvalid 測試多個字段都有無效地址時拒絕發送
func TestIntegrationMultipleRecipientsInvalid(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	viper.Reset()

	host := getTestSMTPHost()
	port := getTestSMTPPort()

	viper.Set("mailField", map[string]any{
		"host":     host,
		"from":     "sender@example.com",
		"to":       "recipient@example.com",
		"cc":       "invalid@",
		"bcc":      "bcc@bad",
		"subject":  "Multiple Invalid Test",
		"contents": "This should not be sent",
		"port":     port,
	})

	ok, err := SendMailWithMultipart("mailField")
	if ok || err == nil {
		t.Fatalf("SendMailWithMultipart should return error, but didn't")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "cc") || !strings.Contains(errMsg, "bcc") {
		t.Fatalf("Error message should contain both 'cc' and 'bcc', got: %v", err)
	}

	t.Logf("✓ Multiple invalid recipients correctly rejected: %v", err)
}

func TestIntegrationStructuredSendToMailpit(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	subject := "Hermes structured 中文 📨"
	body := "Hermes structured 中文 body with emoji 🧪"
	err := NewStructuredSender(NewSMTPMailer()).Send(StructuredSendOptions{
		Server:                  "127.0.0.1",
		Port:                    "1025",
		From:                    "sender@example.com",
		To:                      []string{"to1@example.com", "to2@example.com"},
		CC:                      []string{"cc@example.com"},
		BCC:                     []string{"bcc@example.com"},
		Subject:                 subject,
		Body:                    body,
		ConfirmOutsideWhitelist: true,
	})
	if err != nil {
		t.Fatalf("structured Send() failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	message, err := getLatestMessage()
	if err != nil {
		t.Fatalf("Failed to get latest message: %v", err)
	}
	assertSubjectEquals(t, message, subject)
	assertFromEquals(t, message, "sender@example.com")
	assertToContains(t, message, []string{"to1@example.com", "to2@example.com"})
	assertCcContains(t, message, []string{"cc@example.com"})
	if len(message.Bcc) != 1 || message.Bcc[0].Address != "bcc@example.com" {
		t.Errorf("Mailpit envelope Bcc = %+v, want bcc@example.com", message.Bcc)
	}

	rawMessage, err := getRawMessage(message.ID)
	if err != nil {
		t.Fatalf("Failed to get raw structured message: %v", err)
	}
	// Mailpit reconstructs its raw response with a Bcc header from the SMTP
	// envelope. The pre-SMTP message bytes are covered by unit tests instead.
	if !strings.Contains(rawMessage, "To: to1@example.com,to2@example.com") {
		t.Errorf("raw message does not contain the visible To header")
	}
	if !strings.Contains(rawMessage, "Cc: cc@example.com") {
		t.Errorf("raw message does not contain the visible Cc header")
	}
	assertMIMEStructure(t, rawMessage, "multipart/mixed")
	assertContentContains(t, rawMessage, body)
}

func TestIntegrationStructuredSendMultipleAttachmentsAndFailClosed(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	tempDir := t.TempDir()
	firstPath := tempDir + "/structured-first.txt"
	secondPath := tempDir + "/structured-second.json"
	firstContent := "first structured attachment"
	secondContent := `{"kind":"second structured attachment"}`
	if err := os.WriteFile(firstPath, []byte(firstContent), 0o644); err != nil {
		t.Fatalf("create first attachment: %v", err)
	}
	if err := os.WriteFile(secondPath, []byte(secondContent), 0o644); err != nil {
		t.Fatalf("create second attachment: %v", err)
	}

	sender := NewStructuredSender(NewSMTPMailer())
	options := StructuredSendOptions{
		Server:                  "127.0.0.1",
		Port:                    "1025",
		From:                    "sender@example.com",
		To:                      []string{"recipient@example.com"},
		Subject:                 "Structured multi-attachment",
		Body:                    "attachment body",
		Attachments:             []string{firstPath, secondPath},
		ConfirmOutsideWhitelist: true,
	}
	if err := sender.Send(options); err != nil {
		t.Fatalf("structured attachment Send() failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	message, err := getLatestMessage()
	if err != nil {
		t.Fatalf("Failed to get latest attachment message: %v", err)
	}
	assertAttachmentExists(t, message, "structured-first.txt")
	assertAttachmentExists(t, message, "structured-second.json")
	rawMessage, err := getRawMessage(message.ID)
	if err != nil {
		t.Fatalf("Failed to get raw attachment message: %v", err)
	}
	for _, content := range []string{firstContent, secondContent} {
		if !strings.Contains(rawMessage, base64.StdEncoding.EncodeToString([]byte(content))) {
			t.Errorf("raw message does not contain attachment content %q", content)
		}
	}

	before, err := getMessageCount()
	if err != nil {
		t.Fatalf("Failed to get message count before invalid attachment: %v", err)
	}
	options.Attachments = append(options.Attachments, tempDir+"/missing.txt")
	if err := sender.Send(options); err == nil {
		t.Fatal("structured Send() should reject a missing attachment")
	}
	time.Sleep(100 * time.Millisecond)
	after, err := getMessageCount()
	if err != nil {
		t.Fatalf("Failed to get message count after invalid attachment: %v", err)
	}
	if after != before {
		t.Fatalf("Mailpit count changed after invalid attachment: before=%d after=%d", before, after)
	}
}

func TestIntegrationStructuredHistoryReplayAndFailClosed(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.jsonl")
	firstPath := filepath.Join(tempDir, "history-first.txt")
	secondPath := filepath.Join(tempDir, "history-second.json")
	if err := os.WriteFile(firstPath, []byte("history first attachment"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secondPath, []byte(`{"history":"second"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	before, err := getMessageCount()
	if err != nil {
		t.Fatal(err)
	}
	sender := NewStructuredSender(NewSMTPMailer())
	options := StructuredSendOptions{
		Server:                  "127.0.0.1",
		Port:                    "1025",
		From:                    "sender@example.com",
		To:                      []string{"to@example.com"},
		CC:                      []string{"cc@example.com"},
		BCC:                     []string{"bcc@example.com"},
		Subject:                 "History replay 中文 📨",
		Body:                    "History replay body 中文",
		Attachments:             []string{firstPath, secondPath},
		ConfirmOutsideWhitelist: true,
	}
	if err := SendStructuredWithHistory(sender, options, historyPath); err != nil {
		t.Fatalf("initial history send failed: %v", err)
	}
	records, err := ReadHistory(historyPath)
	if err != nil || len(records) != 1 {
		t.Fatalf("initial history = %+v, %v", records, err)
	}
	replay := ReplayOptions(records[0])
	replay.ConfirmOutsideWhitelist = true
	if err := SendStructuredWithHistory(sender, replay, historyPath); err != nil {
		t.Fatalf("history replay failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	after, err := getMessageCount()
	if err != nil {
		t.Fatal(err)
	}
	if after != before+2 {
		t.Fatalf("Mailpit count after send and replay = %d, want %d", after, before+2)
	}
	message, err := getLatestMessage()
	if err != nil {
		t.Fatal(err)
	}
	assertSubjectEquals(t, message, options.Subject)
	assertToContains(t, message, options.To)
	assertCcContains(t, message, options.CC)
	if len(message.Bcc) != 1 || message.Bcc[0].Address != options.BCC[0] {
		t.Fatalf("replayed Bcc = %+v", message.Bcc)
	}
	assertAttachmentExists(t, message, "history-first.txt")
	assertAttachmentExists(t, message, "history-second.json")
	raw, err := getRawMessage(message.ID)
	if err != nil {
		t.Fatal(err)
	}
	assertContentContains(t, raw, options.Body)

	records, err = ReadHistory(historyPath)
	if err != nil || len(records) != 2 {
		t.Fatalf("history after replay = %+v, %v", records, err)
	}
	stableCount := after
	stableRecords := len(records)

	unconfirmed := ReplayOptions(records[0])
	if err := SendStructuredWithHistory(sender, unconfirmed, historyPath); err == nil {
		t.Fatal("external replay should require fresh confirmation")
	}
	missing := ReplayOptions(records[0])
	missing.ConfirmOutsideWhitelist = true
	missing.Attachments = append(missing.Attachments, filepath.Join(tempDir, "missing.txt"))
	if err := SendStructuredWithHistory(sender, missing, historyPath); err == nil {
		t.Fatal("replay should reject missing attachment")
	}
	damagedPath := filepath.Join(tempDir, "damaged.jsonl")
	if err := os.WriteFile(damagedPath, []byte(`{"version":1`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadHistory(damagedPath); err == nil {
		t.Fatal("damaged history should fail closed")
	}
	time.Sleep(100 * time.Millisecond)
	finalCount, err := getMessageCount()
	if err != nil {
		t.Fatal(err)
	}
	finalRecords, err := ReadHistory(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	if finalCount != stableCount || len(finalRecords) != stableRecords {
		t.Fatalf("fail-closed changed state: mail %d->%d history %d->%d", stableCount, finalCount, stableRecords, len(finalRecords))
	}
}
