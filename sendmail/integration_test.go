// +build integration

package sendmail

import (
	"net/smtp"
	"os"
	"testing"

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
