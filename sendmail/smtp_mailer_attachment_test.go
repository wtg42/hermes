package sendmail

import (
	"encoding/base64"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wtg42/hermes/mail"
)

func TestSMTPMailerSendMultipleAttachments(t *testing.T) {
	dir := t.TempDir()
	firstPath := filepath.Join(dir, "first.txt")
	secondPath := filepath.Join(dir, "second.json")
	firstContent := "first attachment"
	secondContent := `{"name":"second"}`
	if err := os.WriteFile(firstPath, []byte(firstContent), 0o600); err != nil {
		t.Fatalf("建立第一個附件失敗: %v", err)
	}
	if err := os.WriteFile(secondPath, []byte(secondContent), 0o600); err != nil {
		t.Fatalf("建立第二個附件失敗: %v", err)
	}

	original := SendMail
	defer func() { SendMail = original }()

	var gotMessage string
	SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		gotMessage = string(msg)
		return nil
	}

	compose := validAttachmentCompose()
	compose.Attachment = firstPath
	compose.Attachments = []string{firstPath, secondPath, firstPath}

	if err := NewSMTPMailer().Send(compose); err != nil {
		t.Fatalf("SMTPMailer.Send() error = %v", err)
	}

	if count := strings.Count(gotMessage, "Content-Disposition: attachment"); count != 2 {
		t.Fatalf("附件 MIME part 數量 = %d，預期 2", count)
	}
	for _, fileName := range []string{"first.txt", "second.json"} {
		if !strings.Contains(gotMessage, `filename="`+fileName+`"`) {
			t.Errorf("郵件缺少附件 filename %q", fileName)
		}
	}
	for _, content := range []string{firstContent, secondContent} {
		encoded := base64.StdEncoding.EncodeToString([]byte(content))
		if !strings.Contains(gotMessage, encoded) {
			t.Errorf("郵件缺少附件 base64 內容 %q", encoded)
		}
	}
	if !strings.Contains(gotMessage, "Content-Type: text/plain") {
		t.Error("第一個附件缺少 text/plain content type")
	}
	if !strings.Contains(gotMessage, "Content-Type: application/json") {
		t.Error("第二個附件缺少 application/json content type")
	}
}

func TestSMTPMailerSendLegacySingleAttachment(t *testing.T) {
	attachmentPath := filepath.Join(t.TempDir(), "legacy.txt")
	if err := os.WriteFile(attachmentPath, []byte("legacy"), 0o600); err != nil {
		t.Fatalf("建立附件失敗: %v", err)
	}

	original := SendMail
	defer func() { SendMail = original }()

	var gotMessage string
	SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		gotMessage = string(msg)
		return nil
	}

	compose := validAttachmentCompose()
	compose.Attachment = attachmentPath
	if err := NewSMTPMailer().Send(compose); err != nil {
		t.Fatalf("SMTPMailer.Send() error = %v", err)
	}
	if !strings.Contains(gotMessage, `filename="legacy.txt"`) {
		t.Error("舊單一 Attachment 欄位未加入 MIME")
	}
}

func TestSMTPMailerSendRejectsInvalidAttachmentBeforeSMTP(t *testing.T) {
	dir := t.TempDir()
	validPath := filepath.Join(dir, "valid.txt")
	if err := os.WriteFile(validPath, []byte("valid"), 0o600); err != nil {
		t.Fatalf("建立附件失敗: %v", err)
	}

	tests := []struct {
		name        string
		invalidPath string
	}{
		{
			name:        "missing attachment",
			invalidPath: filepath.Join(dir, "missing.txt"),
		},
		{
			name:        "attachment path is a directory",
			invalidPath: dir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := SendMail
			defer func() { SendMail = original }()

			sendCount := 0
			SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
				sendCount++
				return nil
			}

			compose := validAttachmentCompose()
			compose.Attachments = []string{validPath, tt.invalidPath}
			err := NewSMTPMailer().Send(compose)
			if err == nil || !strings.Contains(err.Error(), tt.invalidPath) {
				t.Fatalf("error = %v，預期包含附件路徑 %q", err, tt.invalidPath)
			}
			if sendCount != 0 {
				t.Fatalf("附件驗證失敗仍呼叫 SendMail %d 次", sendCount)
			}
		})
	}
}

func validAttachmentCompose() mail.MailCompose {
	return mail.MailCompose{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Attachment test",
		Body:    "body",
		Host:    "smtp.example.com",
		Port:    "25",
	}
}
