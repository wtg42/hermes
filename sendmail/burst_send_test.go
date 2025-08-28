package sendmail

import (
	"net/smtp"
	"runtime"
	"sync"
	"testing"
	"time"
)

// 測試 BurstModeSendMail 是否依 quantity 併發呼叫 SendMail
func TestBurstModeSendMailConcurrency(t *testing.T) {
	original := SendMail
	defer func() { SendMail = original }()

	var mu sync.Mutex
	total := 0
	current := 0
	maxCurrent := 0

	SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		mu.Lock()
		total++
		current++
		if current > maxCurrent {
			maxCurrent = current
		}
		mu.Unlock()
		time.Sleep(5 * time.Millisecond)
		mu.Lock()
		current--
		mu.Unlock()
		return nil
	}

	qty := runtime.NumCPU() * 2
	BurstModeSendMail(qty, "smtp.example.com", "25", []string{"example.com"})

	if total != qty {
		t.Fatalf("SendMail 呼叫次數預期 %d 次, 實際 %d 次", qty, total)
	}
	if maxCurrent <= 1 {
		t.Fatalf("預期存在併發, 但最大併發度為 %d", maxCurrent)
	}
}
