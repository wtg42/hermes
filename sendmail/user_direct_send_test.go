package sendmail

import (
	"encoding/base64"
	"net/smtp"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// 測試 encodeRFC2047 函數
func TestEncodeRFC2047(t *testing.T) {
	input := "Hello, 世界"
	expected := "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(input)) + "?="
	result := encodeRFC2047(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// 測試 buildEmailHeaders：Cc 應出現在 header
func TestBuildEmailHeaders_CcAppearsInHeader(t *testing.T) {
	data := EmailData{
		From:    "sender@example.com",
		To:      []string{"to@example.com"},
		Cc:      []string{"cc1@example.com", "cc2@example.com"},
		Subject: "test",
	}
	header := buildEmailHeaders(data)

	if !strings.Contains(header, "Cc: cc1@example.com,cc2@example.com") {
		t.Errorf("header 應包含 Cc，實際內容：\n%s", header)
	}
}

// 測試 buildEmailHeaders：Bcc 不應出現在 header
func TestBuildEmailHeaders_BccNotInHeader(t *testing.T) {
	data := EmailData{
		From:    "sender@example.com",
		To:      []string{"to@example.com"},
		Bcc:     []string{"bcc@example.com"},
		Subject: "test",
	}
	header := buildEmailHeaders(data)

	if strings.Contains(header, "Bcc") || strings.Contains(header, "bcc@example.com") {
		t.Errorf("header 不應包含 Bcc，實際內容：\n%s", header)
	}
}

// 測試 buildEmailHeaders：多個 To 以逗號分隔
func TestBuildEmailHeaders_MultipleToJoined(t *testing.T) {
	data := EmailData{
		From:    "sender@example.com",
		To:      []string{"a@example.com", "b@example.com"},
		Subject: "test",
	}
	header := buildEmailHeaders(data)

	if !strings.Contains(header, "To: a@example.com,b@example.com") {
		t.Errorf("header To 欄位應以逗號分隔，實際內容：\n%s", header)
	}
}

// 測試 SendMailWithMultipart：BCC 收件者出現在 SMTP envelope，但不在 header
func TestSendMailWithMultipart_BccInEnvelopeNotHeader(t *testing.T) {
	original := SendMail
	defer func() { SendMail = original }()

	var capturedRecipients []string
	var capturedMsg []byte
	SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		capturedRecipients = to
		capturedMsg = msg
		return nil
	}

	viper.Reset()
	viper.Set("mailField", map[string]any{
		"host":     "localhost",
		"from":     "sender@example.com",
		"to":       "to@example.com",
		"cc":       "",
		"bcc":      "bcc@example.com",
		"subject":  "BCC test",
		"contents": "body",
		"port":     "25",
	})

	ok, err := SendMailWithMultipart("mailField")
	if err != nil || !ok {
		t.Fatalf("SendMailWithMultipart 失敗: %v", err)
	}

	// BCC 應在 SMTP envelope
	found := false
	for _, r := range capturedRecipients {
		if r == "bcc@example.com" {
			found = true
		}
	}
	if !found {
		t.Errorf("BCC 收件者應出現在 SMTP envelope，實際 recipients: %v", capturedRecipients)
	}

	// BCC 不應在 header
	msgStr := string(capturedMsg)
	if strings.Contains(msgStr, "Bcc") || strings.Contains(msgStr, "bcc@example.com") {
		limit := 500
		if len(msgStr) < limit {
			limit = len(msgStr)
		}
		t.Errorf("BCC 不應出現在郵件 header，實際訊息片段：\n%s", msgStr[:limit])
	}
}

// 測試 SendMailWithMultipart：CC 同時在 header 和 SMTP envelope
func TestSendMailWithMultipart_CcInHeaderAndEnvelope(t *testing.T) {
	original := SendMail
	defer func() { SendMail = original }()

	var capturedRecipients []string
	var capturedMsg []byte
	SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		capturedRecipients = to
		capturedMsg = msg
		return nil
	}

	viper.Reset()
	viper.Set("mailField", map[string]any{
		"host":     "localhost",
		"from":     "sender@example.com",
		"to":       "to@example.com",
		"cc":       "cc@example.com",
		"bcc":      "",
		"subject":  "CC test",
		"contents": "body",
		"port":     "25",
	})

	ok, err := SendMailWithMultipart("mailField")
	if err != nil || !ok {
		t.Fatalf("SendMailWithMultipart 失敗: %v", err)
	}

	// CC 應在 SMTP envelope
	found := false
	for _, r := range capturedRecipients {
		if r == "cc@example.com" {
			found = true
		}
	}
	if !found {
		t.Errorf("CC 收件者應出現在 SMTP envelope，實際 recipients: %v", capturedRecipients)
	}

	// CC 應在 header
	msgStr := string(capturedMsg)
	if !strings.Contains(msgStr, "Cc: cc@example.com") {
		limit := 500
		if len(msgStr) < limit {
			limit = len(msgStr)
		}
		t.Errorf("CC 應出現在郵件 header，實際訊息片段：\n%s", msgStr[:limit])
	}
}

// 測試 SendMailWithMultipart：Cc 包含無效地址時函數返回 error
func TestSendMailWithMultipart_CcWithInvalidAddress(t *testing.T) {
	viper.Reset()
	viper.Set("mailField", map[string]any{
		"host":     "localhost",
		"from":     "sender@example.com",
		"to":       "to@example.com",
		"cc":       "invalid@,valid@example.com",
		"bcc":      "",
		"subject":  "CC invalid test",
		"contents": "body",
		"port":     "25",
	})

	ok, err := SendMailWithMultipart("mailField")
	if ok || err == nil {
		t.Fatalf("SendMailWithMultipart 應返回 error，但沒有")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "cc") {
		t.Errorf("錯誤訊息應包含 'cc'，實際：%s", errMsg)
	}
	if !strings.Contains(errMsg, "invalid@") {
		t.Errorf("錯誤訊息應包含無效的地址 'invalid@'，實際：%s", errMsg)
	}
}

// 測試 SendMailWithMultipart：Bcc 包含無效地址時函數返回 error
func TestSendMailWithMultipart_BccWithInvalidAddress(t *testing.T) {
	viper.Reset()
	viper.Set("mailField", map[string]any{
		"host":     "localhost",
		"from":     "sender@example.com",
		"to":       "to@example.com",
		"cc":       "",
		"bcc":      "bcc@invalid",
		"subject":  "BCC invalid test",
		"contents": "body",
		"port":     "25",
	})

	ok, err := SendMailWithMultipart("mailField")
	if ok || err == nil {
		t.Fatalf("SendMailWithMultipart 應返回 error，但沒有")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "bcc") {
		t.Errorf("錯誤訊息應包含 'bcc'，實際：%s", errMsg)
	}
	if !strings.Contains(errMsg, "bcc@invalid") {
		t.Errorf("錯誤訊息應包含無效的地址 'bcc@invalid'，實際：%s", errMsg)
	}
}

// 測試 SendMailWithMultipart：多個字段都有無效地址時一次返回所有錯誤
func TestSendMailWithMultipart_MultipleFieldsWithInvalidAddresses(t *testing.T) {
	viper.Reset()
	viper.Set("mailField", map[string]any{
		"host":     "localhost",
		"from":     "sender@example.com",
		"to":       "to@example.com",
		"cc":       "invalid@",
		"bcc":      "bcc@bad",
		"subject":  "Multiple invalid test",
		"contents": "body",
		"port":     "25",
	})

	ok, err := SendMailWithMultipart("mailField")
	if ok || err == nil {
		t.Fatalf("SendMailWithMultipart 應返回 error，但沒有")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "cc") || !strings.Contains(errMsg, "bcc") {
		t.Errorf("錯誤訊息應同時包含 'cc' 和 'bcc'，實際：%s", errMsg)
	}
}
