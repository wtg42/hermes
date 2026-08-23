package cmd

import (
	"bytes"
	"net/smtp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wtg42/hermes/sendmail"
)

const safeBurstTestDomain = "rd01.softnext.com.tw"

func executeBurstCommand(t *testing.T, args ...string) error {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)

	root := &cobra.Command{
		Use:           "hermes",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(newBurstModeCmd())
	root.SetArgs(append([]string{"burst"}, args...))
	_, err := root.ExecuteC()

	return err
}

func executeBurstCommandWithOutput(t *testing.T, args ...string) (string, error) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)

	output := new(bytes.Buffer)
	root := &cobra.Command{
		Use:           "hermes",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.SetOut(output)
	root.AddCommand(newBurstModeCmd())
	root.SetArgs(append([]string{"burst"}, args...))
	_, err := root.ExecuteC()

	return output.String(), err
}

func TestBurstCmdMissingRequiredFlags(t *testing.T) {
	err := executeBurstCommand(t)
	if err == nil {
		t.Fatal("預期缺少必填旗標時回傳錯誤")
	}
}

func TestBurstCmdRejectsUnsafeInputBeforeSending(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name: "missing random domain",
			args: []string{
				"--quantity", "1",
				"--host", "smtp.example.com",
				"--port", "25",
			},
			wantErr: "--domain",
		},
		{
			name: "unlisted random domain",
			args: []string{
				"--quantity", "1",
				"--host", "smtp.example.com",
				"--port", "25",
				"--domain", "softnext.com.tw",
			},
			wantErr: "softnext.com.tw",
		},
		{
			name: "unlisted fixed recipient",
			args: []string{
				"--quantity", "1",
				"--host", "smtp.example.com",
				"--port", "25",
				"--from", "sender@" + safeBurstTestDomain,
				"--to", "recipient@gmail.com",
			},
			wantErr: "gmail.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := sendmail.SendMail
			defer func() { sendmail.SendMail = original }()

			sendCount := 0
			sendmail.SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
				sendCount++
				return nil
			}

			err := executeBurstCommand(t, tt.args...)
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantErr)) {
				t.Fatalf("error = %v，預期包含 %q", err, tt.wantErr)
			}
			if sendCount != 0 {
				t.Fatalf("驗證失敗仍寄出 %d 封郵件", sendCount)
			}
		})
	}
}

func TestBurstCmdAcceptsSafeAddressModes(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantFrom string
		wantTo   string
	}{
		{
			name: "safe random domain",
			args: []string{
				"--quantity", "1",
				"--host", "smtp.example.com",
				"--port", "25",
				"--domain", safeBurstTestDomain,
			},
		},
		{
			name: "fixed safe addresses",
			args: []string{
				"--quantity", "1",
				"--host", "smtp.example.com",
				"--port", "25",
				"--from", "sender@" + safeBurstTestDomain,
				"--to", "recipient@" + safeBurstTestDomain,
			},
			wantFrom: "sender@" + safeBurstTestDomain,
			wantTo:   "recipient@" + safeBurstTestDomain,
		},
		{
			name: "repeated explicit domain authorization",
			args: []string{
				"--quantity", "1",
				"--host", "smtp.example.com",
				"--port", "25",
				"--from", "sender@softnext.com.tw",
				"--to", "recipient@gmail.com",
				"--allow-domain", "softnext.com.tw",
				"--allow-domain", "gmail.com",
			},
			wantFrom: "sender@softnext.com.tw",
			wantTo:   "recipient@gmail.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := sendmail.SendMail
			defer func() { sendmail.SendMail = original }()

			sendCount := 0
			var gotAddress string
			var gotFrom string
			var gotTo string
			sendmail.SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
				sendCount++
				gotAddress = addr
				gotFrom = from
				gotTo = to[0]
				return nil
			}

			if err := executeBurstCommand(t, tt.args...); err != nil {
				t.Fatalf("執行 burst 命令失敗: %v", err)
			}
			if sendCount != 1 {
				t.Fatalf("SendMail 呼叫次數 = %d，預期 1", sendCount)
			}
			if gotAddress != "smtp.example.com:25" {
				t.Errorf("SMTP address = %q，預期 smtp.example.com:25", gotAddress)
			}
			if tt.wantFrom != "" && gotFrom != tt.wantFrom {
				t.Errorf("From = %q，預期 %q", gotFrom, tt.wantFrom)
			}
			if tt.wantFrom == "" && !strings.HasSuffix(gotFrom, "@"+safeBurstTestDomain) {
				t.Errorf("隨機 From = %q，預期使用安全網域", gotFrom)
			}
			if tt.wantTo != "" && gotTo != tt.wantTo {
				t.Errorf("To = %q，預期 %q", gotTo, tt.wantTo)
			}
			if tt.wantTo == "" && !strings.HasSuffix(gotTo, "@"+safeBurstTestDomain) {
				t.Errorf("隨機 To = %q，預期使用安全網域", gotTo)
			}
		})
	}
}

func TestBurstCmdDiagnosticFlagsAndSummary(t *testing.T) {
	original := sendmail.SendMail
	defer func() { sendmail.SendMail = original }()

	var message string
	sendmail.SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		message = string(msg)
		return nil
	}

	output, err := executeBurstCommandWithOutput(t,
		"--quantity", "1",
		"--host", "smtp.example.com",
		"--port", "25",
		"--from", "sender@"+safeBurstTestDomain,
		"--to", "recipient@"+safeBurstTestDomain,
		"--run-id", "backup-20230719",
		"--subject-prefix", "MSE-BACKUP",
		"--body-kb", "2",
		"--workers", "1",
		"--rate", "0",
	)
	if err != nil {
		t.Fatalf("執行 burst 命令失敗: %v", err)
	}
	if !strings.Contains(message, "Message-ID: <hermes-backup-20230719-000001@"+safeBurstTestDomain+">") {
		t.Fatalf("郵件缺少指定 run-id 的 Message-ID")
	}
	if len(message) < 2*1024 {
		t.Fatalf("郵件 size = %d，預期至少 2 KiB", len(message))
	}
	if !strings.Contains(output, "run=backup-20230719") || !strings.Contains(output, "succeeded=1") || !strings.Contains(output, "failed=0") {
		t.Fatalf("summary output = %q", output)
	}
}

func TestBurstCmdRequiresBulkConfirmation(t *testing.T) {
	original := sendmail.SendMail
	defer func() { sendmail.SendMail = original }()

	sendCount := 0
	sendmail.SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		sendCount++
		return nil
	}
	err := executeBurstCommand(t,
		"--quantity", "1001",
		"--host", "smtp.example.com",
		"--port", "25",
		"--from", "sender@"+safeBurstTestDomain,
		"--to", "recipient@"+safeBurstTestDomain,
	)
	if err == nil || !strings.Contains(err.Error(), "confirm-burst") {
		t.Fatalf("error = %v，預期要求 --confirm-burst", err)
	}
	if sendCount != 0 {
		t.Fatalf("未確認大量寄送仍寄出 %d 封", sendCount)
	}
}

func TestBurstCmdAsyncDoesNotSendSMTPDirectly(t *testing.T) {
	original := sendmail.SendMail
	defer func() { sendmail.SendMail = original }()

	sendCount := 0
	sendmail.SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		sendCount++
		return nil
	}
	err := executeBurstCommand(t,
		"--quantity", "1",
		"--host", "smtp.example.com",
		"--port", "25",
		"--from", "sender@"+safeBurstTestDomain,
		"--to", "recipient@"+safeBurstTestDomain,
		"--async",
		"--queue-url", "amqp://127.0.0.1:1/",
	)
	if err == nil || !strings.Contains(err.Error(), "connect to RabbitMQ") {
		t.Fatalf("error = %v, want RabbitMQ connection error", err)
	}
	if sendCount != 0 {
		t.Fatalf("direct SendMail calls = %d, want 0", sendCount)
	}
}
