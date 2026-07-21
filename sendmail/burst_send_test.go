package sendmail

import (
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
