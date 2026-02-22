//go:build integration
// +build integration

package sendmail

import (
	"net/smtp"
	"os"
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
	receiverDomains := []string{"example.com", "test.com"}
	BurstModeSendMail(quantity, host, port, receiverDomains)

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
