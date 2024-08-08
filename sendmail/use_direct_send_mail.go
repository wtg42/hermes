package sendmail

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
)

// 郵件主旨需要使用 base64 編碼來解決中文編碼問題
func encodeRFC2047(String string) string {
	// 使用 base64 編碼
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(String)) + "?="
}

// public function
// 目前呼叫 DirectSendMail() 函數來發送郵件
func DirectSendMail() {
	from := "weitingshih@rd01.softnext.com.tw"
	// password := "yourpassword"
	to := "weitingshih@softnext.com.tw"
	subject := "Subject: 測試郵件主旨\r\n"
	body := "白卜庭: Good, good! Let the hate flow through you."

	// 設置 MIME 標頭
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
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

	msg += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	// 設定 SMTP 伺服器資訊
	smtpHost := "192.168.91.61"
	smtpPort := "25"

	// auth := smtp.PlainAuth("", from, password, smtpHost)
	fmt.Println(msg)
	err := smtp.SendMail(smtpHost+":"+smtpPort, nil, from, []string{to}, []byte(msg))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Email sent successfully")
}
