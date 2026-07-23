package sendmail

import (
	"errors"
	"fmt"
	"net/smtp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

const safeBurstDomain = "rd01.softnext.com.tw"

type capturedBurstMail struct {
	from string
	to   string
	msg  string
}

func TestBurstModeSendMailAddressModes(t *testing.T) {
	tests := []struct {
		name       string
		options    BurstOptions
		wantFrom   string
		wantTo     string
		randomFrom bool
		randomTo   bool
	}{
		{
			name: "fixed from and fixed to",
			options: BurstOptions{
				Quantity: 2,
				Host:     "smtp.example.com",
				Port:     "25",
				From:     "sender@" + safeBurstDomain,
				To:       "recipient@" + safeBurstDomain,
			},
			wantFrom: "sender@" + safeBurstDomain,
			wantTo:   "recipient@" + safeBurstDomain,
		},
		{
			name: "fixed from and random to",
			options: BurstOptions{
				Quantity: 2,
				Host:     "smtp.example.com",
				Port:     "25",
				From:     "sender@" + safeBurstDomain,
				Domains:  []string{safeBurstDomain},
			},
			wantFrom: "sender@" + safeBurstDomain,
			randomTo: true,
		},
		{
			name: "random from and fixed to",
			options: BurstOptions{
				Quantity: 2,
				Host:     "smtp.example.com",
				Port:     "25",
				To:       "recipient@" + safeBurstDomain,
				Domains:  []string{safeBurstDomain},
			},
			randomFrom: true,
			wantTo:     "recipient@" + safeBurstDomain,
		},
		{
			name: "random from and random to",
			options: BurstOptions{
				Quantity: 2,
				Host:     "smtp.example.com",
				Port:     "25",
				Domains:  []string{safeBurstDomain},
			},
			randomFrom: true,
			randomTo:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := SendMail
			defer func() { SendMail = original }()

			var mu sync.Mutex
			captured := make([]capturedBurstMail, 0, tt.options.Quantity)
			SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
				mu.Lock()
				defer mu.Unlock()
				captured = append(captured, capturedBurstMail{
					from: from,
					to:   to[0],
					msg:  string(msg),
				})
				return nil
			}

			if err := BurstModeSendMail(tt.options); err != nil {
				t.Fatalf("BurstModeSendMail() error = %v", err)
			}
			if len(captured) != tt.options.Quantity {
				t.Fatalf("SendMail 呼叫次數預期 %d 次，實際 %d 次", tt.options.Quantity, len(captured))
			}

			for _, mail := range captured {
				if tt.randomFrom {
					if !strings.HasSuffix(mail.from, "@"+safeBurstDomain) {
						t.Errorf("隨機 From = %q，預期使用 %s", mail.from, safeBurstDomain)
					}
				} else if mail.from != tt.wantFrom {
					t.Errorf("From = %q，預期 %q", mail.from, tt.wantFrom)
				}

				if tt.randomTo {
					if !strings.HasSuffix(mail.to, "@"+safeBurstDomain) {
						t.Errorf("隨機 To = %q，預期使用 %s", mail.to, safeBurstDomain)
					}
				} else if mail.to != tt.wantTo {
					t.Errorf("To = %q，預期 %q", mail.to, tt.wantTo)
				}

				if !strings.Contains(mail.msg, "From: "+mail.from+"\r\n") {
					t.Errorf("郵件 header 未包含 SMTP envelope From %q", mail.from)
				}
				if !strings.Contains(mail.msg, "To: "+mail.to+"\r\n") {
					t.Errorf("郵件 header 未包含 SMTP envelope To %q", mail.to)
				}
			}
		})
	}
}

func TestBurstModeSendMailDomainAuthorization(t *testing.T) {
	tests := []struct {
		name        string
		options     BurstOptions
		wantErr     string
		wantSendCnt int
	}{
		{
			name: "built-in safe domain",
			options: BurstOptions{
				Quantity: 1,
				Host:     "smtp.example.com",
				Port:     "25",
				Domains:  []string{safeBurstDomain},
			},
			wantSendCnt: 1,
		},
		{
			name: "unlisted domain rejected",
			options: BurstOptions{
				Quantity: 1,
				Host:     "smtp.example.com",
				Port:     "25",
				Domains:  []string{"softnext.com.tw"},
			},
			wantErr: "softnext.com.tw",
		},
		{
			name: "exact authorization accepted",
			options: BurstOptions{
				Quantity:       1,
				Host:           "smtp.example.com",
				Port:           "25",
				Domains:        []string{"softnext.com.tw"},
				AllowedDomains: []string{"softnext.com.tw"},
			},
			wantSendCnt: 1,
		},
		{
			name: "parent does not authorize subdomain",
			options: BurstOptions{
				Quantity:       1,
				Host:           "smtp.example.com",
				Port:           "25",
				Domains:        []string{"mail.softnext.com.tw"},
				AllowedDomains: []string{"softnext.com.tw"},
			},
			wantErr: "mail.softnext.com.tw",
		},
		{
			name: "all domains must be authorized",
			options: BurstOptions{
				Quantity:       1,
				Host:           "smtp.example.com",
				Port:           "25",
				Domains:        []string{safeBurstDomain, "example.com", "gmail.com"},
				AllowedDomains: []string{"example.com"},
			},
			wantErr: "gmail.com",
		},
		{
			name: "fixed recipient domain requires authorization",
			options: BurstOptions{
				Quantity: 1,
				Host:     "smtp.example.com",
				Port:     "25",
				From:     "sender@" + safeBurstDomain,
				To:       "recipient@gmail.com",
			},
			wantErr: "gmail.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := SendMail
			defer func() { SendMail = original }()

			sendCount := 0
			SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
				sendCount++
				return nil
			}

			err := BurstModeSendMail(tt.options)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v，預期包含 %q", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("BurstModeSendMail() error = %v", err)
			}
			if sendCount != tt.wantSendCnt {
				t.Fatalf("SendMail 呼叫次數 = %d，預期 %d", sendCount, tt.wantSendCnt)
			}
		})
	}
}

func TestBurstModeSendMailPreflightValidation(t *testing.T) {
	tests := []struct {
		name    string
		options BurstOptions
		wantErr string
	}{
		{
			name: "missing random domain",
			options: BurstOptions{
				Quantity: 1,
				Host:     "smtp.example.com",
				Port:     "25",
				To:       "recipient@" + safeBurstDomain,
			},
			wantErr: "domain",
		},
		{
			name: "invalid fixed from",
			options: BurstOptions{
				Quantity: 1,
				Host:     "smtp.example.com",
				Port:     "25",
				From:     "invalid@",
				To:       "recipient@" + safeBurstDomain,
			},
			wantErr: "from",
		},
		{
			name: "invalid random domain",
			options: BurstOptions{
				Quantity: 1,
				Host:     "smtp.example.com",
				Port:     "25",
				Domains:  []string{"bad_domain"},
			},
			wantErr: "bad_domain",
		},
		{
			name: "invalid allowed domain",
			options: BurstOptions{
				Quantity:       1,
				Host:           "smtp.example.com",
				Port:           "25",
				Domains:        []string{safeBurstDomain},
				AllowedDomains: []string{"bad_domain"},
			},
			wantErr: "bad_domain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := SendMail
			defer func() { SendMail = original }()

			sendCount := 0
			SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
				sendCount++
				return nil
			}

			err := BurstModeSendMail(tt.options)
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantErr)) {
				t.Fatalf("error = %v，預期包含 %q", err, tt.wantErr)
			}
			if sendCount != 0 {
				t.Fatalf("預先驗證失敗仍呼叫 SendMail %d 次", sendCount)
			}
		})
	}
}

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
	if err := BurstModeSendMail(BurstOptions{
		Quantity: qty,
		Host:     "smtp.example.com",
		Port:     "25",
		Domains:  []string{safeBurstDomain},
	}); err != nil {
		t.Fatalf("BurstModeSendMail() error = %v", err)
	}

	if total != qty {
		t.Fatalf("SendMail 呼叫次數預期 %d 次, 實際 %d 次", qty, total)
	}
	if maxCurrent <= 1 {
		t.Fatalf("預期存在併發, 但最大併發度為 %d", maxCurrent)
	}
}

func TestExecuteBurstBuildsTraceableMessages(t *testing.T) {
	original := SendMail
	defer func() { SendMail = original }()

	var captured []string
	SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		captured = append(captured, string(msg))
		return nil
	}

	result, err := ExecuteBurst(BurstOptions{
		Quantity:      3,
		Host:          "smtp.example.com",
		Port:          "25",
		From:          "sender@" + safeBurstDomain,
		To:            "recipient@" + safeBurstDomain,
		RunID:         "backup-20230719",
		SubjectPrefix: "MSE-BACKUP",
		BodyKB:        2,
		Workers:       1,
	})
	if err != nil {
		t.Fatalf("ExecuteBurst() error = %v", err)
	}
	if result.RunID != "backup-20230719" || result.Requested != 3 || result.Attempted != 3 || result.Succeeded != 3 || result.Failed != 0 {
		t.Fatalf("ExecuteBurst() result = %+v", result)
	}
	if len(captured) != 3 {
		t.Fatalf("captured messages = %d，預期 3", len(captured))
	}

	for index, message := range captured {
		sequence := index + 1
		subject := fmt.Sprintf("MSE-BACKUP-backup-20230719-%06d", sequence)
		messageID := fmt.Sprintf("Message-ID: <hermes-backup-20230719-%06d@%s>\r\n", sequence, safeBurstDomain)
		if !strings.Contains(message, "Subject: "+encodeRFC2047(subject)+"\r\n") {
			t.Errorf("message %d 缺少可追蹤 Subject", sequence)
		}
		if !strings.Contains(message, messageID) {
			t.Errorf("message %d 缺少可追蹤 Message-ID", sequence)
		}
		if !strings.Contains(message, "Date: ") {
			t.Errorf("message %d 缺少 Date header", sequence)
		}
		if len(message) < 2*1024 {
			t.Errorf("message %d size = %d，預期至少 2 KiB", sequence, len(message))
		}
	}
}

func TestExecuteBurstReturnsCompleteStatistics(t *testing.T) {
	original := SendMail
	defer func() { SendMail = original }()

	call := 0
	SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		call++
		if call == 2 {
			return errors.New("SMTP rejected message")
		}
		return nil
	}

	var progress []BurstProgress
	result, err := ExecuteBurst(BurstOptions{
		Quantity: 3,
		Host:     "smtp.example.com",
		Port:     "25",
		From:     "sender@" + safeBurstDomain,
		To:       "recipient@" + safeBurstDomain,
		RunID:    "stats-run",
		Workers:  1,
		Progress: func(update BurstProgress) {
			progress = append(progress, update)
		},
	})
	if err == nil || !strings.Contains(err.Error(), "1 failed") {
		t.Fatalf("ExecuteBurst() error = %v，預期包含失敗統計", err)
	}
	if result.Attempted != 3 || result.Succeeded != 2 || result.Failed != 1 {
		t.Fatalf("ExecuteBurst() result = %+v", result)
	}
	if len(progress) == 0 || progress[len(progress)-1].Attempted != 3 {
		t.Fatalf("最後 progress = %+v，預期 attempted=3", progress)
	}
}

func TestBurstModeSendMailRequiresConfirmationAboveBulkLimit(t *testing.T) {
	original := SendMail
	defer func() { SendMail = original }()

	sendCount := 0
	SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		sendCount++
		return nil
	}

	err := BurstModeSendMail(BurstOptions{
		Quantity: 1001,
		Host:     "smtp.example.com",
		Port:     "25",
		From:     "sender@" + safeBurstDomain,
		To:       "recipient@" + safeBurstDomain,
	})
	if err == nil || !strings.Contains(err.Error(), "confirm-burst") {
		t.Fatalf("error = %v，預期要求 --confirm-burst", err)
	}
	if sendCount != 0 {
		t.Fatalf("未確認大量寄送仍寄出 %d 封", sendCount)
	}
}

func TestBuildBurstPlanValidatesDiagnosticControls(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*BurstOptions)
		wantErr string
	}{
		{name: "invalid run id", mutate: func(options *BurstOptions) { options.RunID = "bad run" }, wantErr: "run-id"},
		{name: "header injection", mutate: func(options *BurstOptions) { options.SubjectPrefix = "safe\r\nBcc: victim@example.com" }, wantErr: "subject-prefix"},
		{name: "negative body size", mutate: func(options *BurstOptions) { options.BodyKB = -1 }, wantErr: "body-kb"},
		{name: "too many workers", mutate: func(options *BurstOptions) { options.Workers = 257 }, wantErr: "workers"},
		{name: "excessive rate", mutate: func(options *BurstOptions) { options.RatePerSecond = 10001 }, wantErr: "rate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := BurstOptions{
				Quantity: 1,
				Host:     "smtp.example.com",
				Port:     "25",
				From:     "sender@" + safeBurstDomain,
				To:       "recipient@" + safeBurstDomain,
			}
			tt.mutate(&options)

			_, err := buildBurstPlan(options)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("buildBurstPlan() error = %v，預期包含 %q", err, tt.wantErr)
			}
		})
	}
}

func TestBurstInterval(t *testing.T) {
	if got := burstInterval(250); got != 4*time.Millisecond {
		t.Fatalf("burstInterval(250) = %s，預期 4ms", got)
	}
	if got := burstInterval(0); got != 0 {
		t.Fatalf("burstInterval(0) = %s，預期不限制", got)
	}
}
