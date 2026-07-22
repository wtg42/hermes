// sendmail package 為底層發信函數或信件處理函數
// 利用 CLI 或是 TUI 收集的參數來發信
// 統一使用 SendMailWithMultipart() 函數進行郵件發送
package sendmail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"strings"

	"github.com/spf13/viper"
	"github.com/wtg42/hermes/mail"
	"github.com/wtg42/hermes/utils"
)

// 郵件主旨需要使用 base64 編碼來解決中文編碼問題
func encodeRFC2047(String string) string {
	// 使用 base64 編碼
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(String)) + "?="
}

// SendMailFunc 為 smtp.SendMail 的函式類型
//   - 便於測試時進行依賴注入
type SendMailFunc func(addr string, a smtp.Auth, from string, to []string, msg []byte) error

// EmailData 包含郵件的所有必要信息
type EmailData struct {
	Host     string
	Port     string
	From     string
	To       []string
	Cc       []string
	Bcc      []string
	Subject  string
	Contents string
}

// SMTPMailer 實作 mail.Mailer 介面，使用 SMTP 發送郵件
type SMTPMailer struct{}

// NewSMTPMailer 建立新的 SMTPMailer 實例
func NewSMTPMailer() *SMTPMailer {
	return &SMTPMailer{}
}

// Send 透過 SMTP 發送郵件，實作 mail.Mailer 介面
func (m *SMTPMailer) Send(compose mail.MailCompose) error {
	return m.SendWithTransport(compose, DefaultSMTPTransportConfig(compose.Host, compose.Port))
}

// SendWithTransport sends one message using an explicit SMTP transport policy.
func (m *SMTPMailer) SendWithTransport(compose mail.MailCompose, transport SMTPTransportConfig) error {
	// 驗證必要欄位
	toEmails, invalidTo := utils.ValidateEmails(strings.Join(compose.To, ","))
	if len(toEmails) == 0 {
		return fmt.Errorf("no valid 'to' addresses")
	}

	ccEmails, invalidCc := utils.ValidateEmails(strings.Join(compose.CC, ","))
	if len(strings.Join(compose.CC, ",")) == 0 {
		invalidCc = nil
	}

	bccEmails, invalidBcc := utils.ValidateEmails(strings.Join(compose.BCC, ","))
	if len(strings.Join(compose.BCC, ",")) == 0 {
		invalidBcc = nil
	}

	// 檢查 Cc/Bcc 是否有無效地址
	var errMsgs []string
	if len(invalidTo) > 0 {
		errMsgs = append(errMsgs, fmt.Sprintf("invalid addresses in 'to': %v", invalidTo))
	}
	if len(invalidCc) > 0 {
		errMsgs = append(errMsgs, fmt.Sprintf("invalid addresses in 'cc': %v", invalidCc))
	}
	if len(invalidBcc) > 0 {
		errMsgs = append(errMsgs, fmt.Sprintf("invalid addresses in 'bcc': %v", invalidBcc))
	}

	if len(errMsgs) > 0 {
		return fmt.Errorf("%s", strings.Join(errMsgs, "; "))
	}

	attachmentPaths := normalizeAttachmentPaths(compose.Attachment, compose.Attachments)
	attachments, err := loadAttachments(attachmentPaths)
	if err != nil {
		return err
	}

	// 構建郵件
	email := new(bytes.Buffer)

	// 寫入 header
	headers := make(map[string]string)
	headers["From"] = compose.From
	if len(toEmails) > 0 {
		headers["To"] = strings.Join(toEmails, ",")
	}
	if len(ccEmails) > 0 {
		headers["Cc"] = strings.Join(ccEmails, ",")
	}
	headers["Subject"] = encodeRFC2047(compose.Subject)

	for k, v := range headers {
		email.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	// 構建 MIME content
	if err := m.buildMIMEContent(email, compose.Body, attachments); err != nil {
		return fmt.Errorf("failed to build MIME content: %w", err)
	}

	// 準備所有收件者（to, cc, bcc）
	allRecipients := append([]string{}, toEmails...)
	allRecipients = append(allRecipients, ccEmails...)
	allRecipients = append(allRecipients, bccEmails...)

	// 發送郵件
	if err := sendSMTPWithTransport(transport, compose.From, allRecipients, email.Bytes()); err != nil {
		return fmt.Errorf("failed to send mail: %w", err)
	}

	log.Println("Email sent successfully")
	return nil
}

// buildMIMEContent 構建 MIME multipart 郵件內容
//   - email: 目標 buffer，用於寫入 MIME 內容
//   - body: 郵件正文
//   - attachments: 已完成驗證與載入的附件
//   - 返回 error 如果內容構建失敗
func (m *SMTPMailer) buildMIMEContent(email *bytes.Buffer, body string, attachments []*Attachment) error {
	return buildMIMEContentWithAttachments(email, body, attachments)
}

func normalizeAttachmentPaths(legacyPath string, paths []string) []string {
	ordered := make([]string, 0, len(paths)+1)
	seen := make(map[string]struct{}, len(paths)+1)
	appendPath := func(path string) {
		if _, exists := seen[path]; exists {
			return
		}
		seen[path] = struct{}{}
		ordered = append(ordered, path)
	}

	if legacyPath != "" {
		appendPath(legacyPath)
	}
	for _, path := range paths {
		appendPath(path)
	}

	return ordered
}

func loadAttachments(paths []string) ([]*Attachment, error) {
	attachments := make([]*Attachment, 0, len(paths))
	for _, path := range paths {
		attachment, err := NewAttachment(path)
		if err != nil {
			return nil, fmt.Errorf("attachment %q: %w", path, err)
		}
		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

func buildMIMEContentWithAttachments(email *bytes.Buffer, body string, attachments []*Attachment) error {
	writer := multipart.NewWriter(email)

	contentType := fmt.Sprintf("multipart/mixed; boundary=%s;", writer.Boundary())
	fmt.Fprintf(email, "Content-Type: %s\r\n", contentType)
	fmt.Fprintf(email, "MIME-Version: 1.0\r\n\r\n")

	// 文字內容部分
	{
		partHead := textproto.MIMEHeader{
			"Content-Type":              {"text/plain; charset=\"utf-8\""},
			"Content-Transfer-Encoding": {"base64"},
		}
		part, err := writer.CreatePart(partHead)
		if err != nil {
			return fmt.Errorf("failed to create text part: %w", err)
		}
		// base64 編碼以支援中文
		if _, err := part.Write([]byte(base64.StdEncoding.EncodeToString([]byte(body)))); err != nil {
			return fmt.Errorf("failed to write text part: %w", err)
		}
	}

	for _, attachment := range attachments {
		partAttachHead := textproto.MIMEHeader{}
		partAttachHead.Set("Content-Type", attachment.ContentType)
		partAttachHead.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachment.FileName))
		partAttachHead.Set("Content-Transfer-Encoding", attachment.Encoding)

		part, err := writer.CreatePart(partAttachHead)
		if err != nil {
			return fmt.Errorf("attachment %q: failed to create MIME part: %w", attachment.FilePath, err)
		}
		if _, err := part.Write([]byte(attachment.EncodedFile)); err != nil {
			return fmt.Errorf("attachment %q: failed to write MIME part: %w", attachment.FilePath, err)
		}
	}

	// HTML 內容部分
	{
		part, err := writer.CreatePart(map[string][]string{"Content-Type": {"text/html"}})
		if err != nil {
			return fmt.Errorf("failed to create HTML part: %w", err)
		}
		if _, err := part.Write([]byte("<html><body><h1>" + body + "</h1></body></html>")); err != nil {
			return fmt.Errorf("failed to write HTML part: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close MIME content: %w", err)
	}

	return nil
}

// buildEmailHeaders 構建郵件 header 部分（為向後相容保留）
//   - data: 郵件信息
//   - 返回 header 字符串（不包含 body）
func buildEmailHeaders(data EmailData) string {
	headers := make(map[string]string)
	headers["From"] = data.From

	if len(data.To) > 0 {
		headers["To"] = strings.Join(data.To, ",")
	}
	if len(data.Cc) > 0 {
		headers["Cc"] = strings.Join(data.Cc, ",")
	}

	headers["Subject"] = encodeRFC2047(data.Subject)

	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	return msg
}

// buildMIMEContent 構建 MIME multipart 郵件內容（為向後相容保留）
//   - email: 目標 buffer，用於寫入 MIME 內容
//   - contents: 郵件正文
//   - 返回 error 如果內容構建失敗
func buildMIMEContent(email *bytes.Buffer, contents string) error {
	return buildMIMEContentWithAttachments(email, contents, nil)
}

func legacyMailCompose(fields map[string]any) (mail.MailCompose, error) {
	required := make(map[string]string, 5)
	for _, field := range []string{"host", "from", "to", "subject", "contents"} {
		value, ok := fields[field].(string)
		if !ok {
			return mail.MailCompose{}, fmt.Errorf("%s configuration missing or invalid", field)
		}
		required[field] = value
	}

	optional := make(map[string]string, 4)
	for _, field := range []string{"port", "cc", "bcc", "attachment"} {
		value, exists := fields[field]
		if !exists {
			continue
		}
		stringValue, ok := value.(string)
		if !ok {
			return mail.MailCompose{}, fmt.Errorf("%s configuration invalid", field)
		}
		optional[field] = stringValue
	}

	port := optional["port"]
	if port == "" {
		port = "25"
	}

	return mail.MailCompose{
		Host:       required["host"],
		Port:       port,
		From:       required["from"],
		To:         splitLegacyAddresses(required["to"]),
		CC:         splitLegacyAddresses(optional["cc"]),
		BCC:        splitLegacyAddresses(optional["bcc"]),
		Subject:    required["subject"],
		Body:       required["contents"],
		Attachment: optional["attachment"],
	}, nil
}

func splitLegacyAddresses(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
}

// SendMailWithMultipart adapts the legacy Viper map to the common SMTP pipeline.
func SendMailWithMultipart(key string) (bool, error) {
	compose, err := legacyMailCompose(viper.GetStringMap(key))
	if err != nil {
		return false, err
	}
	if err := NewSMTPMailer().Send(compose); err != nil {
		return false, err
	}
	return true, nil
}

// SendMail 包裝 smtp.SendMail 方便測試
//   - addr: SMTP 伺服器位址
//   - from: 寄件者
//   - to: 收件者清單
//   - msg: 原始郵件內容
//
// 使用變數讓測試可以注入假的 SendMail 行為
var SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	return smtp.SendMail(addr, nil, from, to, msg)
}
