package sendmail

import (
	"net/smtp"
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
