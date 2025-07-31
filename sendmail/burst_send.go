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
	// 步驟 1: 準備郵件發送池
	// 為了避免發送太多重複的電子郵件，我們創建一個郵件地址池。
	// 這個池會預先生成一定數量的隨機電子郵件地址。
	const totalEmailNum int = 100 // 設定郵件池的大小
	mailPool := make([]string, totalEmailNum)
	for i := range mailPool {
		// 為郵件池中的每個位置生成一個隨機的電子郵件地址
		mailPool[i] = utils.RandomEmail(receiverDomain)
	}

	// 這個函數做很簡單的事情 就是依照 total 數量發送隨機產生的郵件
	// 步驟 2: 定義發送單封郵件的匿名函數
	// 這個函數負責根據給定的總數發送隨機生成的郵件。
	doSendEmails := func(total int) {
		// 初始化一個新的隨機數生成器，確保每次調用都有不同的隨機序列
		var source = rand.NewSource(time.Now().UnixNano())
		var r = rand.New(source)

		msg := strings.Builder{}
		// 迴圈指定次數，每次構建並發送一封郵件
		for range make([]struct{}, total) {
			// 從郵件池中隨機選擇發件人和收件人
			from := mailPool[r.Intn(len(mailPool))]
			to := mailPool[r.Intn(len(mailPool))]

			// 設置郵件的 MIME 標頭，包括發件人、收件人、和主題
			headers := make(map[string]string)
			headers["From"] = from
			headers["To"] = to
			headers["Subject"] = encodeRFC2047(utils.RandomString(10))

			// 將標頭寫入郵件內容構建器
			for k, v := range headers {
				msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
			}

			// 創建 multipart 寫入器，用於處理郵件的多部分內容（例如：文本、HTML、附件）
			multipartWriter := multipart.NewWriter(&msg)
			contentType := fmt.Sprintf("multipart/mixed; boundary=%s;", multipartWriter.Boundary())
			msg.WriteString(fmt.Sprintf("Content-Type: %s\r\n", contentType))
			msg.WriteString("MIME-Version: 1.0\r\n\r\n") // 加入 MIME-Version 標頭

			// 創建文本部分的標頭
			textPart := textproto.MIMEHeader{
				"Content-Type":              {"text/plain; charset=\"utf-8\""},
				"Content-Transfer-Encoding": {"base64"},
			}
			// 創建文本部分並寫入內容
			part, err := multipartWriter.CreatePart(textPart)
			if err != nil {
				log.Println("CreatePart Error:", err)
			}
			// 將郵件內容進行 base64 編碼，以支援中文和其他特殊字符
			part.Write([]byte(base64.StdEncoding.EncodeToString([]byte(utils.RandomString(50)))))

			{
				// 創建 HTML 部分的標頭
				part, err := multipartWriter.CreatePart(map[string][]string{"Content-Type": {"text/html"}})
				if err != nil {
					panic(err)
				}

				// 寫入 HTML 內容
				part.Write([]byte("<html><body><h1>Sent by Hermes</h1></body></html>"))
			}

			// 關閉 multipart 寫入器，完成郵件內容的構建
			err = multipartWriter.Close()
			if err != nil {
				log.Println("Close Error:", err)
			}

			// 發送郵件
			SendMail(host+":"+port, nil, from, []string{to}, []byte(msg.String()))
			// err = smtp.SendMail(host+":"+port, nil, from, []string{to}, []byte(msg.String()))
			if err != nil {
				log.Println("Error:", err)
			}

			// 重置郵件內容構建器，為下一封郵件做準備，避免內容混淆
			msg.Reset()
		}
	}

	// 步驟 3: 分配任務給多個 Goroutine
	// 將總郵件發送數量分配給系統可用的 CPU 核心數，以實現併發發送。
	numGoroutine := runtime.NumCPU() // 獲取 CPU 核心數
	// 計算每個核心需要處理的任務數量
	tasksPerCore := quantity / numGoroutine
	// 計算剩餘的任務數量，這些任務將分配給第一個 Goroutine
	remainder := quantity % numGoroutine

	// 創建一個 channel 用於接收 Goroutine 的完成訊息
	ch := make(chan string, 1)
	// 創建一個 WaitGroup 用於等待所有 Goroutine 完成
	var wg sync.WaitGroup
	// 啟動多個 Goroutine 併發發送郵件
	for i := 0; i < numGoroutine; i++ {
		wg.Add(1) // 增加 WaitGroup 計數器
		// 複製迴圈變數 i 的值，以避免 Goroutine 閉包問題
		index := i
		go func() {
			defer wg.Done() // Goroutine 完成時減少 WaitGroup 計數器
			// 第一個 Goroutine 需要處理額外的剩餘任務
			if index == 0 && remainder > 0 {
				doSendEmails(int(tasksPerCore) + int(remainder))
			} else {
				// 其他 Goroutine 處理平均分配的任務數量
				doSendEmails(int(tasksPerCore))
			}
			ch <- "mail sent." // 發送完成訊息到 channel
		}()

		// 如果每個核心的任務數量為 0，表示所有任務已分配完畢，提前結束迴圈
		if tasksPerCore == 0 {
			break
		}
	}

	// 步驟 4: 啟動一個 Goroutine 接收並列印發送結果
	go func() {
		for v := range ch {
			log.Println(v)
		}
	}()

	// 步驟 5: 等待所有 Goroutine 完成
	// 阻塞直到所有發送郵件的 Goroutine 都執行完畢
	wg.Wait()
	// 關閉 channel
	close(ch)
}
