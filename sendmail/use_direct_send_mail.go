package sendmail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"mime/multipart"
	"net/smtp"
	"net/textproto"

	"github.com/samber/lo"
	"github.com/spf13/viper"
)

// 郵件主旨需要使用 base64 編碼來解決中文編碼問題
func encodeRFC2047(String string) string {
	// 使用 base64 編碼
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(String)) + "?="
}

// public function
// 目前呼叫 DirectSendMail() 函數來發送郵件
func DirectSendMail() {
	// 使用用戶的輸入設定郵件
	host := viper.GetString("host")
	from := viper.GetString("from")
	to := viper.GetString("to")
	// password := "yourpassword"
	subject := viper.GetString("subject") + "\r\n"
	body := viper.GetString("body")

	// 設置 MIME 標頭
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Cc"] = "weiting.shi1982@gmail.com"
	headers["Subject"] = encodeRFC2047(subject)
	headers["MIME-Version"] = "1.0"
	// 設定 utf-8
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""
	// 設定 base64 編碼
	headers["Content-Transfer-Encoding"] = "base64"

	// 構建郵件內容
	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	// 將郵件內容進行 base64 編碼 才能支援中文
	msg += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	// 設定 SMTP 伺服器資訊
	smtpHost := host
	smtpPort := "25"

	// auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, nil, from, []string{to}, []byte(msg))
	if err != nil {
		log.Println("Error:", err)
		return
	}
	log.Println("Email sent successfully")
}

// 基本的文字訊息
func DirectSendMailFromTui(key string) (bool, error) {
	if !lo.Contains([]string{"mailField"}, key) {
		return false, fmt.Errorf("key %v 不在範圍內", key)
	}

	// 使用用戶的輸入設定郵件
	mailFields := viper.GetStringMap(key)

	host := mailFields["host"].(string)
	from := mailFields["from"].(string)
	to := mailFields["to"].(string)
	cc := mailFields["cc"].(string)
	bcc := mailFields["bcc"].(string)
	// password := "yourpassword"
	subject := mailFields["subject"].(string) + "\r\n"
	body := mailFields["contents"].(string)

	// 設置 MIME 標頭
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Cc"] = cc
	headers["Bcc"] = bcc
	headers["Subject"] = encodeRFC2047(subject)
	headers["MIME-Version"] = "1.0"
	// 設定 utf-8
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""
	// 設定 base64 編碼
	headers["Content-Transfer-Encoding"] = "base64"

	// 構建郵件內容
	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	// 將郵件內容進行 base64 編碼 才能支援中文
	msg += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	// 設定 SMTP 伺服器資訊
	smtpHost := host
	smtpPort := "1025"

	// auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, nil, from, []string{to}, []byte(msg))
	if err != nil {
		log.Println("Error:", err)
		return false, err
	}
	log.Println("Email sent successfully")

	return true, nil
}

// 有支援 multipart 版本
func SendMailWithMultipart(key string) (bool, error) {
	if !lo.Contains([]string{"mailField"}, key) {
		return false, fmt.Errorf("key %v 不在範圍內", key)
	}
	// you need a pointer to bytes.Buffer
	email := new(bytes.Buffer)

	// 使用用戶的輸入設定郵件
	mailFields := viper.GetStringMap(key)

	host := mailFields["host"].(string)
	from := mailFields["from"].(string)
	to := mailFields["to"].(string)
	cc := mailFields["cc"].(string)
	bcc := mailFields["bcc"].(string)
	subject := mailFields["subject"].(string)
	body := mailFields["contents"].(string)

	// 設置 MIME 標頭
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Cc"] = cc
	headers["Bcc"] = bcc
	headers["Subject"] = encodeRFC2047(subject)

	for k, v := range headers {
		email.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	// header 寫完後就可以換 multipart 最開始部分
	writer := multipart.NewWriter(email)
	defer writer.Close()

	contentType := fmt.Sprintf("multipart/mixed; boundary=%s;", writer.Boundary())
	fmt.Fprintf(email, "Content-Type: %s\r\n", contentType)
	fmt.Fprintf(email, "MIME-Version: 1.0\r\n\r\n") // 加入 MIME-Version

	// 創建 MIEM 部分 CreatePart() 會返回這個 part 的 writer(自動處理邊界跟內容)
	// 另外為了對應中文部分要用 base64 編碼
	partHead := textproto.MIMEHeader{
		"Content-Type":              {"text/plain; charset=\"utf-8\""},
		"Content-Transfer-Encoding": {"base64"},
	}
	part, err := writer.CreatePart(partHead)
	if err != nil {
		log.Println("Error:", err)
		return false, err
	}

	// 將郵件內容進行 base64 編碼 才能支援中文
	part.Write([]byte(base64.StdEncoding.EncodeToString([]byte(body))))

	// 創建另一個部分，設定為 HTML 內容
	part, err = writer.CreatePart(map[string][]string{"Content-Type": {"text/html"}})
	if err != nil {
		panic(err)
	}

	part.Write([]byte("<html><body><h1>Hello, World!</h1></body></html>"))

	// 在所有部分都寫入後，關閉 writer 以添加結束邊界
	err = writer.Close()
	if err != nil {
		log.Println("Error closing writer:", err)
		return false, err
	}

	// 設定 SMTP 伺服器資訊
	smtpHost := host
	smtpPort := "1025"

	// auth := smtp.PlainAuth("", from, password, smtpHost)
	err = smtp.SendMail(smtpHost+":"+smtpPort, nil, from, []string{to}, email.Bytes())
	if err != nil {
		log.Println("Error:", err)
		return false, err
	}
	log.Println("Email sent successfully")

	return true, nil
}
