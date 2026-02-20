package sendmail

import (
	"net/smtp"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// 測試 SendMailWithMultipart 在有附件時會包含附件標頭
func TestSendMailWithMultipartWithAttachment(t *testing.T) {
	viper.Reset()

	tmp, err := os.CreateTemp("", "attach*.txt")
	if err != nil {
		t.Fatalf("建立暫存檔失敗: %v", err)
	}
	defer os.Remove(tmp.Name())
	tmp.WriteString("hello")
	tmp.Close()

	viper.Set("mailField", map[string]any{
		"host":       "smtp.example.com",
		"from":       "from@example.com",
		"to":         "to@example.com",
		"cc":         "",
		"bcc":        "",
		"subject":    "hi",
		"contents":   "body",
		"port":       "25",
		"attachment": tmp.Name(),
	})

	original := SendMail
	defer func() { SendMail = original }()

	var gotMsg []byte
	SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		gotMsg = append([]byte{}, msg...)
		return nil
	}

	ok, err := SendMailWithMultipart("mailField")
	if err != nil || !ok {
		t.Fatalf("SendMailWithMultipart 失敗: %v", err)
	}
	if !strings.Contains(string(gotMsg), "Content-Disposition: attachment") {
		t.Errorf("應包含附件標頭, got %s", string(gotMsg))
	}
}

// 測試 SendMailWithMultipart 在沒有附件時不包含附件標頭
func TestSendMailWithMultipartWithoutAttachment(t *testing.T) {
	viper.Reset()
	viper.Set("mailField", map[string]any{
		"host":     "smtp.example.com",
		"from":     "from@example.com",
		"to":       "to@example.com",
		"cc":       "",
		"bcc":      "",
		"subject":  "hi",
		"contents": "body",
		"port":     "25",
	})

	original := SendMail
	defer func() { SendMail = original }()

	var gotMsg []byte
	SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		gotMsg = append([]byte{}, msg...)
		return nil
	}

	ok, err := SendMailWithMultipart("mailField")
	if err != nil || !ok {
		t.Fatalf("SendMailWithMultipart 失敗: %v", err)
	}
	if strings.Contains(string(gotMsg), "Content-Disposition: attachment") {
		t.Errorf("不應包含附件標頭, got %s", string(gotMsg))
	}
}

// 測試 SendMailWithMultipart 支援 BCC
func TestSendMailWithMultipartWithBcc(t *testing.T) {
	viper.Reset()
	viper.Set("mailField", map[string]any{
		"host":     "smtp.example.com",
		"from":     "from@example.com",
		"to":       "to@example.com",
		"cc":       "cc@example.com",
		"bcc":      "bcc@example.com",
		"subject":  "hi",
		"contents": "body",
		"port":     "25",
	})

	original := SendMail
	defer func() { SendMail = original }()

	var gotRecipients []string
	var gotMsg []byte
	SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		gotRecipients = append([]string{}, to...)
		gotMsg = append([]byte{}, msg...)
		return nil
	}

	ok, err := SendMailWithMultipart("mailField")
	if err != nil || !ok {
		t.Fatalf("SendMailWithMultipart 失敗: %v", err)
	}

	// Should have 3 recipients: to, cc, bcc
	if len(gotRecipients) != 3 {
		t.Errorf("收件人應包含 to, cc, bcc, got %v", gotRecipients)
	}

	// BCC should not appear in headers
	if strings.Contains(string(gotMsg), "Bcc:") {
		t.Errorf("不應包含 Bcc 標頭 (應為 blind copy), got %s", string(gotMsg))
	}

	// CC should appear in headers
	if !strings.Contains(string(gotMsg), "Cc: cc@example.com") {
		t.Errorf("應包含 Cc 標頭, got %s", string(gotMsg))
	}
}
