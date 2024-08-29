package sendmail

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"time"

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

func DirectSendMailFromTui(key string) bool {
	if !lo.Contains([]string{"mailField"}, key) {
		return false
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
	smtpPort := "25"

	// auth := smtp.PlainAuth("", from, password, smtpHost)
	time.Sleep(3 * time.Second)
	err := smtp.SendMail(smtpHost+":"+smtpPort, nil, from, []string{to}, []byte(msg))
	if err != nil {
		log.Println("Error:", err)
		return false
	}
	log.Println("Email sent successfully")

	return true
}
