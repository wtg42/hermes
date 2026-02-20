package sendmail

import (
	"bytes"
	"log"
	"runtime"
	"sync"
	"time"

	"math/rand"

	"github.com/wtg42/hermes/utils"
)

// BurstModeSendMail 瘋狂發送郵件 - 用於壓力測試
//   - quantity: 需要發送的郵件數量
//   - host: SMTP 主機名稱
//   - port: SMTP 通訊埠
//   - receiverDomain: 收件者網域清單
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

	// 步驟 2: 定義發送單封郵件的匿名函數
	// 這個函數負責根據給定的總數發送隨機生成的郵件
	doSendEmails := func(total int) {
		// 初始化一個新的隨機數生成器，確保每次調用都有不同的隨機序列
		var source = rand.NewSource(time.Now().UnixNano())
		var r = rand.New(source)

		// 迴圈指定次數，每次構建並發送一封郵件
		for range make([]struct{}, total) {
			// 從郵件池中隨機選擇發件人和收件人
			from := mailPool[r.Intn(len(mailPool))]
			to := mailPool[r.Intn(len(mailPool))]

			// 構建郵件信息
			data := EmailData{
				Host:     host,
				Port:     port,
				From:     from,
				To:       []string{to},
				Cc:       []string{},
				Bcc:      []string{},
				Subject:  utils.RandomString(10),
				Contents: utils.RandomString(50),
			}

			// 構建郵件
			email := new(bytes.Buffer)
			headerStr := buildEmailHeaders(data)
			email.WriteString(headerStr)

			// 構建 MIME content
			err := buildMIMEContent(email, data.Contents)
			if err != nil {
				log.Printf("Error building MIME content: %v\n", err)
				continue
			}

			// 發送郵件
			err = SendMail(host+":"+port, nil, from, []string{to}, email.Bytes())
			if err != nil {
				log.Printf("Error sending mail: %v\n", err)
				continue
			}
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
