package main

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
)

func main() {
	directSendMail()
	// basicSendMail()
}

func directSendMail() {
	from := "weitingshih@softnext.com.tw"
	// password := "yourpassword"
	to := "weitingshih@rd01.softnext.com.tw"
	subject := "Subject: 測試郵件主旨\r\n"
	body := "這是郵件內容。"

	// 設置 MIME 標頭
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = encodeRFC2047(subject)
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""
	headers["Content-Transfer-Encoding"] = "base64"

	// 構建郵件內容
	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + body

	// 設定 SMTP 伺服器資訊
	smtpHost := "192.168.91.61"
	smtpPort := "25"

	// auth := smtp.PlainAuth("", from, password, smtpHost)
	// fmt.Println([]byte(msg))
	err := smtp.SendMail(smtpHost+":"+smtpPort, nil, from, []string{to}, []byte(msg))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Email sent successfully")
}

// 郵件主旨需要使用 base64 編碼來解決中文編碼問題
func encodeRFC2047(String string) string {
	// 使用 base64 編碼
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(String)) + "?="
}

func basicSendMail() {
	smtpHost := "192.168.91.61"
	smtpPort := "25"

	// 設定收件人、寄件人及信件內容
	from := "weitingshih@softnext.com.tw"
	to := []string{"weitingshih@rd01.softnext.com.tw"}

	// msg := []byte("To: recipient@example.com\r\n" +
	// 	"Subject: 測試郵件\r\n" +
	// 	"\r\n" +
	// 	"這是一封測試郵件。\r\n")
	msg := []byte("To: recipient@example.com")

	// 建立未加密的 SMTP 客戶端連接
	c, err := smtp.Dial(smtpHost + ":" + smtpPort)
	if err != nil {
		fmt.Printf("無法連接到 SMTP 伺服器: %s\n", err)
		return
	}
	defer c.Close()

	// 認證
	// if err = c.Auth(auth); err != nil {
	// 	fmt.Printf("認證失敗: %s\n", err)
	// 	return
	// }

	// 設置寄件人
	if err = c.Mail(from); err != nil {
		fmt.Printf("寄件人設置失敗: %s\n", err)
		return
	}

	// 設置收件人
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			fmt.Printf("收件人設置失敗: %s\n", err)
			return
		}
	}

	// 寄送資料
	wc, err := c.Data()
	if err != nil {
		fmt.Printf("無法取得寫入接口: %s\n", err)
		return
	}

	_, err = wc.Write(msg)
	if err != nil {
		fmt.Printf("寫入資料失敗: %s\n", err)
		return
	}

	err = wc.Close()
	if err != nil {
		fmt.Printf("無法關閉寫入接口: %s\n", err)
		return
	}

	fmt.Println("郵件寄送成功!")
}
