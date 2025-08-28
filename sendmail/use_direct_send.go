// sendmail package 為底層發信函數或信件處理函數
// 利用 CLI 或是 TUI 收集的參數來發信
// 依據不同功能來呼叫不同的函數來發信
// TODO: 重構改用 Interface
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

	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/wtg42/hermes/utils"
)

// 郵件主旨需要使用 base64 編碼來解決中文編碼問題
func encodeRFC2047(String string) string {
	// 使用 base64 編碼
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(String)) + "?="
}

// 這只是為了測試用改成 DI 方式 但其實把 smtp.SendMail 額外包裝成一個函數就好了
type SendMailFunc func(addr string, a smtp.Auth, from string, to []string, msg []byte) error

// 目前呼叫 DirectSendMail() 函數來發送郵件
// 僅純文字郵件發送
func DirectSendMail(s SendMailFunc) {
	// 使用用戶的輸入設定郵件
	host := viper.GetString("host")
	port := viper.GetString("port")
	from := viper.GetString("from")
	to := viper.GetString("to")
	// password := "yourpassword"
	subject := viper.GetString("subject") + "\r\n"
	contents := viper.GetString("contents")

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
	msg += "\r\n" + base64.StdEncoding.EncodeToString([]byte(contents))

	// 設定 SMTP 伺服器資訊
	// auth := smtp.PlainAuth("", from, password, smtpHost)
	err := s(host+":"+port, nil, from, []string{to}, []byte(msg))
	if err != nil {
		log.Println("Error:", err)
		return
	}
	log.Println("Email sent successfully")
}

// Deprecated: Use SendMailWithMultipart instead
//
// 基本的文字訊息郵件發送版本 目前被 mutilipart 版本代替
func DirectSendMailFromTui(key string) (bool, error) {
	if !lo.Contains([]string{"mailField"}, key) {
		return false, fmt.Errorf("key %v 不在範圍內", key)
	}

	// 使用用戶的輸入設定郵件
	mailFields := viper.GetStringMap(key)

	host := mailFields["host"].(string)
	port := mailFields["port"].(string)
	from := mailFields["from"].(string)
	to := mailFields["to"].(string)
	cc := mailFields["cc"].(string)
	// bcc := mailFields["bcc"].(string)
	// password := "yourpassword"
	subject := mailFields["subject"].(string) + "\r\n"
	body := mailFields["contents"].(string)

	// 設置 MIME 標頭
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Cc"] = cc
	// headers["Bcc"] = bcc
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
	smtpPort := port

	// auth := smtp.PlainAuth("", from, password, smtpHost)
	err := SendMail(smtpHost+":"+smtpPort, nil, from, []string{to}, []byte(msg))
	if err != nil {
		log.Println("Error:", err)
		return false, err
	}
	log.Println("Email sent successfully")

	return true, nil
}

// 有支援 multipart 版本發信
// 取代 DirectSendMailFromTui()
func SendMailWithMultipart(key string) (bool, error) {
	if !lo.Contains([]string{"mailField"}, key) {
		return false, fmt.Errorf("key %v 不在範圍內", key)
	}
	// you need a pointer to bytes.Buffer
	email := new(bytes.Buffer)

	// 使用 viper 資料庫取得用戶的輸入設定郵件
	mailFields := viper.GetStringMap(key)

	host := mailFields["host"].(string)
	from := mailFields["from"].(string)

	toEmails, _ := utils.ValidateEmails(mailFields["to"].(string))
	to := strings.Join(toEmails, ",")
	log.Println("tttttt=>", to)

	ccEmails, _ := utils.ValidateEmails(mailFields["cc"].(string))
	cc := strings.Join(ccEmails, ",")

	// bccEmails, _ := utils.ValidateEmails(mailFields["bcc"].(string))
	// bcc := strings.Join(bccEmails, ",")

	subject := mailFields["subject"].(string)
	contents := mailFields["contents"].(string)
	port := mailFields["port"].(string)
	if port == "" {
		port = "25"
	}

	{
		// 設置 MIME 標頭
		headers := make(map[string]string)
		headers["From"] = from
		headers["To"] = to
		headers["Cc"] = cc
		// headers["Bcc"] = bcc
		headers["Subject"] = encodeRFC2047(subject)

		for k, v := range headers {
			email.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}

	// Header 跟 Body 部分都有各自的寫入規格 觀念不要混在一起了 重構時候要注意
	// Header 寫完後就可以換 multipart 最開始部分
	writer := multipart.NewWriter(email)
	defer writer.Close()

	contentType := fmt.Sprintf("multipart/mixed; boundary=%s;", writer.Boundary())
	fmt.Fprintf(email, "Content-Type: %s\r\n", contentType)
	fmt.Fprintf(email, "MIME-Version: 1.0\r\n\r\n") // 加入 MIME-Version

	{
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
		part.Write([]byte(base64.StdEncoding.EncodeToString([]byte(contents))))
	}

	{
		// 附件夾檔部分 若用戶沒給檔案或是無效則跳過
		attachment := Attachment{}
		ok, err := attachment.NewAttachment()
		if err != nil {
			log.Println("Error:", err)
			return false, err
		}
		if ok {
			partAttachHead := textproto.MIMEHeader{}
			partAttachHead.Set("Content-Type", attachment.ContentType)
			partAttachHead.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachment.FileName))
			partAttachHead.Set("Content-Transfer-Encoding", attachment.Encoding)

			part, err := writer.CreatePart(partAttachHead)
			if err != nil {
				log.Fatalln("Error:", err)
			}
			part.Write([]byte(attachment.EncodedFile))
		}
	}

	{
		// 創建另一個部分，設定為 HTML 內容
		part, err := writer.CreatePart(map[string][]string{"Content-Type": {"text/html"}})
		if err != nil {
			panic(err)
		}

		part.Write([]byte("<html><body><h1>" + contents + "</h1></body></html>"))
	}

	// 在所有部分都寫入後，關閉 writer 以添加結束邊界
	err := writer.Close()
	if err != nil {
		log.Println("Error closing writer:", err)
		return false, err
	}

	// 設定 SMTP 伺服器資訊
	smtpHost := host
	smtpPort := port

	// time.Sleep(1 * time.Second)
	err = SendMail(smtpHost+":"+smtpPort, nil, from, []string{to}, email.Bytes())
	if err != nil {
		log.Println("Error:", err)
		return false, err
	}
	log.Println("Email sent successfully")

	return true, nil
}

// 一切都只是為了好測試才把這個包裝起來
func SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	return smtp.SendMail(addr, nil, from, to, msg)
}
