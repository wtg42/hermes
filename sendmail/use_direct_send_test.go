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
