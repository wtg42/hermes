package sendmail

import (
	"net/smtp"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// 測試沒有 cc 時收件者與訊息
func TestDirectSendMailWithoutCc(t *testing.T) {
	viper.Reset()
	viper.Set("host", "smtp.example.com")
	viper.Set("port", "25")
	viper.Set("from", "from@example.com")
	viper.Set("to", "to@example.com")
	viper.Set("subject", "test")
	viper.Set("contents", "body")

	var gotRecipients []string
	var gotMsg string
	mock := func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		gotRecipients = append([]string{}, to...)
		gotMsg = string(msg)
		return nil
	}

	DirectSendMail(mock)

	if len(gotRecipients) != 1 || gotRecipients[0] != "to@example.com" {
		t.Errorf("收件人應只有 to, got %v", gotRecipients)
	}
	if strings.Contains(gotMsg, "Cc:") {
		t.Errorf("不應包含 Cc 標頭")
	}
}

// 測試有 cc 時收件者與訊息
func TestDirectSendMailWithCc(t *testing.T) {
	viper.Reset()
	viper.Set("host", "smtp.example.com")
	viper.Set("port", "25")
	viper.Set("from", "from@example.com")
	viper.Set("to", "to@example.com")
	viper.Set("cc", "cc@example.com")
	viper.Set("subject", "test")
	viper.Set("contents", "body")

	var gotRecipients []string
	var gotMsg string
	mock := func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		gotRecipients = append([]string{}, to...)
		gotMsg = string(msg)
		return nil
	}

	DirectSendMail(mock)

	if len(gotRecipients) != 2 {
		t.Errorf("收件人應包含 to 與 cc, got %v", gotRecipients)
	}
	if !strings.Contains(gotMsg, "Cc: cc@example.com") {
		t.Errorf("應包含 Cc 標頭, got %s", gotMsg)
	}
}

// 測試有 bcc 時收件者與訊息
func TestDirectSendMailWithBcc(t *testing.T) {
	viper.Reset()
	viper.Set("host", "smtp.example.com")
	viper.Set("port", "25")
	viper.Set("from", "from@example.com")
	viper.Set("to", "to@example.com")
	viper.Set("bcc", "bcc@example.com")
	viper.Set("subject", "test")
	viper.Set("contents", "body")

	var gotRecipients []string
	var gotMsg string
	mock := func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		gotRecipients = append([]string{}, to...)
		gotMsg = string(msg)
		return nil
	}

	DirectSendMail(mock)

	if len(gotRecipients) != 2 {
		t.Errorf("收件人應包含 to 與 bcc, got %v", gotRecipients)
	}
	if strings.Contains(gotMsg, "Bcc:") {
		t.Errorf("不應包含 Bcc 標頭 (應為 blind copy), got %s", gotMsg)
	}
}

// 測試同時有 cc 與 bcc 時收件者與訊息
func TestDirectSendMailWithCcAndBcc(t *testing.T) {
	viper.Reset()
	viper.Set("host", "smtp.example.com")
	viper.Set("port", "25")
	viper.Set("from", "from@example.com")
	viper.Set("to", "to@example.com,to2@example.com")
	viper.Set("cc", "cc@example.com")
	viper.Set("bcc", "bcc@example.com,bcc2@example.com")
	viper.Set("subject", "test")
	viper.Set("contents", "body")

	var gotRecipients []string
	var gotMsg string
	mock := func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		gotRecipients = append([]string{}, to...)
		gotMsg = string(msg)
		return nil
	}

	DirectSendMail(mock)

	// Should have: to@example.com, to2@example.com, cc@example.com, bcc@example.com, bcc2@example.com
	if len(gotRecipients) != 5 {
		t.Errorf("收件人應包含所有 to, cc, bcc 地址, got %v", gotRecipients)
	}
	if !strings.Contains(gotMsg, "Cc: cc@example.com") {
		t.Errorf("應包含 Cc 標頭, got %s", gotMsg)
	}
	if strings.Contains(gotMsg, "Bcc:") {
		t.Errorf("不應包含 Bcc 標頭 (應為 blind copy), got %s", gotMsg)
	}
}

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
