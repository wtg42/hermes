package sendmail

import (
	"encoding/base64"
	"fmt"
	"log"
	"mime/multipart"
	"net/textproto"
	"runtime"
	"strings"
	"sync"
	"time"

	"math/rand"

	"github.com/wtg42/hermes/utils"
)

// 瘋狂發送郵件
//   - quantity: 需要發送的郵件數量
//   - host: smtp 主機名稱
//   - port: smtp ports
//   - receiverDomain: 需要發送郵件的 email 網域名
func BurstModeSendMail(quantity int, host string, port string, receiverDomain []string) {
	// 至少要有一定的數量 才不會發生太多重複 email
	const totalEmailNum int = 100
	mailPool := make([]string, totalEmailNum)
	for i := range mailPool {
		mailPool[i] = utils.RandomEmail(receiverDomain)
	}

	// 這個函數做很簡單的事情 就是依照 total 數量發送隨機產生的郵件
	doSendEmails := func(total int) {
		// 另外一個 random seed
		var source = rand.NewSource(time.Now().UnixNano())
		var r = rand.New(source)

		msg := strings.Builder{}
		for i := 0; i < total; i++ {
			from := mailPool[r.Intn(len(mailPool))]
			to := mailPool[r.Intn(len(mailPool))]

			// 設置 MIME 標頭
			headers := make(map[string]string)
			headers["From"] = from
			headers["To"] = to
			headers["Subject"] = encodeRFC2047(utils.RandomString(10))

			// 構建郵件內容
			for k, v := range headers {
				msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
			}

			multipartWriter := multipart.NewWriter(&msg)
			contentType := fmt.Sprintf("multipart/mixed; boundary=%s;", multipartWriter.Boundary())
			msg.WriteString(fmt.Sprintf("Content-Type: %s\r\n", contentType))
			msg.WriteString("MIME-Version: 1.0\r\n\r\n") // 加入 MIME-Version

			textPart := textproto.MIMEHeader{
				"Content-Type":              {"text/plain; charset=\"utf-8\""},
				"Content-Transfer-Encoding": {"base64"},
			}
			part, err := multipartWriter.CreatePart(textPart)
			if err != nil {
				log.Println("CreatePart Error:", err)
			}
			// 將郵件內容進行 base64 編碼 才能支援中文
			part.Write([]byte(base64.StdEncoding.EncodeToString([]byte(utils.RandomString(50)))))

			{
				// 創建另一個部分，設定為 HTML 內容
				part, err := multipartWriter.CreatePart(map[string][]string{"Content-Type": {"text/html"}})
				if err != nil {
					panic(err)
				}

				part.Write([]byte("<html><body><h1>Sent by Hermes</h1></body></html>"))
			}

			err = multipartWriter.Close()
			if err != nil {
				log.Println("Close Error:", err)
			}

			SendMail(host+":"+port, nil, from, []string{to}, []byte(msg.String()))
			// err = smtp.SendMail(host+":"+port, nil, from, []string{to}, []byte(msg.String()))
			if err != nil {
				log.Println("Error:", err)
			}

			// 重置 msg 一定要做 不然會有奇怪的隨機性 bcc 欄位寫入
			msg.Reset()
		}
	}

	// 把總數量分配給系統可以用的 CPU 核心數
	numGoroutine := runtime.NumCPU()
	// the number of tasks for each core
	tasksPerCore := quantity / numGoroutine
	// the remainder of tasks
	remainder := quantity % numGoroutine

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

		// 0 表示不用用到全部數量 CPU 跑第一次就可以把跑完了
		if tasksPerCore == 0 {
			break
		}
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
