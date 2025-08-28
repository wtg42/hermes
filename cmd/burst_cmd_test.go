package cmd

import (
	"net/smtp"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wtg42/hermes/sendmail"
)

// TestBurstCmdMissingRequiredFlags 確認缺少必填旗標時回傳錯誤
func TestBurstCmdMissingRequiredFlags(t *testing.T) {
	viper.Reset()
	root := &cobra.Command{Use: "hermes"}
	root.AddCommand(burstModeCmd)
	// 重新綁定旗標到 viper，避免 Reset 後遺失設定
	viper.BindPFlag("burst-quantity", burstModeCmd.PersistentFlags().Lookup("quantity"))
	viper.BindPFlag("burst-host", burstModeCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("burst-port", burstModeCmd.PersistentFlags().Lookup("port"))

	root.SetArgs([]string{"burst"})
	_, err := root.ExecuteC()
	if err == nil {
		t.Fatalf("預期缺少旗標時應回傳錯誤")
	}
}

// TestBurstCmdRunCalled 確認提供旗標時 Run 會被執行
func TestBurstCmdRunCalled(t *testing.T) {
	viper.Reset()
	root := &cobra.Command{Use: "hermes"}
	root.AddCommand(burstModeCmd)
	// 重新綁定旗標到 viper，避免 Reset 後遺失設定
	viper.BindPFlag("burst-quantity", burstModeCmd.PersistentFlags().Lookup("quantity"))
	viper.BindPFlag("burst-host", burstModeCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("burst-port", burstModeCmd.PersistentFlags().Lookup("port"))

	called := false
	original := sendmail.SendMail
	sendmail.SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		called = true
		return nil
	}
	defer func() { sendmail.SendMail = original }()

	root.SetArgs([]string{"burst", "--quantity", "1", "--host", "smtp.example.com", "--port", "25"})
	if _, err := root.ExecuteC(); err != nil {
		t.Fatalf("執行命令應無錯誤: %v", err)
	}
	if !called {
		t.Fatalf("預期 SendMail 被呼叫")
	}
}
