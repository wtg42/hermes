package sendmail

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"runtime"
	"sync"
	"time"

	"math/rand"

	"github.com/wtg42/hermes/utils"
)

// 瘋狂發送郵件
func BurstModeSendMail(quantity int, host string, port string) {
	// 依照用戶輸入的數量生成等量的隨機 email
	mailPool := make([]string, quantity)
	for i := range mailPool {
		mailPool[i] = utils.RandomEmail()
	}

	// 把總數量分配給系統可以用的 CPU 核心數
	numGoroutine := runtime.NumCPU()
	// the number of tasks for each core
	tasksPerCore := quantity / numGoroutine
	// the remainder of tasks
	remainder := quantity % numGoroutine

	// 這個函數做很簡單的事情 就是依照 total 數量發送隨機產生的郵件
	doSendEmails := func(total int) {
		// 另外一個 random seed
		source := rand.NewSource(time.Now().UnixNano())
		r := rand.New(source)

		for i := 0; i < total; i++ {
			from := mailPool[r.Intn(len(mailPool))]
			to := mailPool[r.Intn(len(mailPool))]

			// 設置 MIME 標頭
			headers := make(map[string]string)
			headers["From"] = from
			headers["To"] = to
			headers["Cc"] = ""
			headers["Bcc"] = ""
			headers["Subject"] = encodeRFC2047(utils.RandomString(10))
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
			msg += "\r\n" + base64.StdEncoding.EncodeToString([]byte(utils.RandomString(50)))

			err := smtp.SendMail(host+":"+port, nil, from, []string{to}, []byte(msg))
			if err != nil {
				log.Println("Error:", err)
			}
		}
	}

	ch := make(chan string, 1)
	var wg sync.WaitGroup
	for i := 0; i < numGoroutine; i++ {
		wg.Add(1)

		// Go 1.23.2 編譯器檢查似乎不允許 goroutine 使用共享的 i 變數
		index := i
		go func() {
			defer wg.Done()
			if index == 0 && remainder > 0 {
				// 第一個 goroutine 需要再多處理餘數部分
				doSendEmails(int(tasksPerCore) + int(remainder))
			} else {
				doSendEmails(int(tasksPerCore))
			}
			ch <- "mail sent."
		}()
	}

	// 啟動接收結果的 goroutine
	go func() {
		for v := range ch {
			log.Println(v)
		}
	}()

	wg.Wait()
	close(ch)
}

func GenerateNumberOfEmails(amount int) []string {
	emails := make([]string, amount)
	for i := range emails {
		emails[i] = utils.RandomEmail()
	}
	return emails
}
