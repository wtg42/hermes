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

// buildEmailHeaders 構建郵件 header 部分
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

// buildMIMEContent 構建 MIME multipart 郵件內容
//   - email: 目標 buffer，用於寫入 MIME 內容
//   - contents: 郵件正文
//   - 返回 error 如果內容構建失敗
func buildMIMEContent(email *bytes.Buffer, contents string) error {
	writer := multipart.NewWriter(email)
	defer writer.Close()

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
		part.Write([]byte(base64.StdEncoding.EncodeToString([]byte(contents))))
	}

	// 附件部分 - 失敗時只記錄警告，不中斷郵件發送
	{
		attachment := Attachment{}
		ok, err := attachment.NewAttachment()
		if err != nil {
			log.Printf("Warning: failed to process attachment: %v\n", err)
		} else if ok {
			partAttachHead := textproto.MIMEHeader{}
			partAttachHead.Set("Content-Type", attachment.ContentType)
			partAttachHead.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachment.FileName))
			partAttachHead.Set("Content-Transfer-Encoding", attachment.Encoding)

			part, err := writer.CreatePart(partAttachHead)
			if err != nil {
				log.Printf("Warning: failed to create attachment part: %v\n", err)
			} else {
				part.Write([]byte(attachment.EncodedFile))
			}
		}
	}

	// HTML 內容部分
	{
		part, err := writer.CreatePart(map[string][]string{"Content-Type": {"text/html"}})
		if err != nil {
			return fmt.Errorf("failed to create HTML part: %w", err)
		}
		part.Write([]byte("<html><body><h1>" + contents + "</h1></body></html>"))
	}

	return nil
}

// SendMailWithMultipart 以 MIME multipart 格式發送郵件
//   - key: 從 viper 取得設定的鍵值
//   - 支援純文字、HTML 內容及附件
//   - 返回 (成功標記, 錯誤)
func SendMailWithMultipart(key string) (bool, error) {
	// 取得郵件配置
	mailFields := viper.GetStringMap(key)

	host, ok := mailFields["host"].(string)
	if !ok {
		return false, fmt.Errorf("host configuration missing or invalid")
	}

	from, ok := mailFields["from"].(string)
	if !ok {
		return false, fmt.Errorf("from configuration missing or invalid")
	}

	toEmails, invalidTo := utils.ValidateEmails(mailFields["to"].(string))
	if len(toEmails) == 0 {
		return false, fmt.Errorf("no valid 'to' addresses")
	}

	ccEmails, _ := utils.ValidateEmails(mailFields["cc"].(string))
	bccEmails, _ := utils.ValidateEmails(mailFields["bcc"].(string))

	subject, ok := mailFields["subject"].(string)
	if !ok {
		return false, fmt.Errorf("subject configuration missing or invalid")
	}

	contents, ok := mailFields["contents"].(string)
	if !ok {
		return false, fmt.Errorf("contents configuration missing or invalid")
	}

	port, ok := mailFields["port"].(string)
	if !ok || port == "" {
		port = "25"
	}

	// 警告用戶如果有無效的 email
	if len(invalidTo) > 0 {
		log.Printf("Warning: Invalid 'to' addresses filtered out: %v\n", invalidTo)
	}

	// 構建郵件
	email := new(bytes.Buffer)

	// 寫入 header
	data := EmailData{
		Host:     host,
		Port:     port,
		From:     from,
		To:       toEmails,
		Cc:       ccEmails,
		Subject:  subject,
		Contents: contents,
	}

	headerStr := buildEmailHeaders(data)
	email.WriteString(headerStr)

	// 構建 MIME content
	if err := buildMIMEContent(email, contents); err != nil {
		return false, fmt.Errorf("failed to build MIME content: %w", err)
	}

	// 準備所有收件者（to, cc, bcc）
	allRecipients := append([]string{}, toEmails...)
	allRecipients = append(allRecipients, ccEmails...)
	allRecipients = append(allRecipients, bccEmails...)

	// 發送郵件
	err := SendMail(host+":"+port, nil, from, allRecipients, email.Bytes())
	if err != nil {
		return false, fmt.Errorf("failed to send mail: %w", err)
	}

	log.Println("Email sent successfully")
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
