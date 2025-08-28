package cmd

import (
	"net/smtp"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wtg42/hermes/sendmail"
)

// TestDirectSendCmdMissingRequiredFlags 確認缺少必填旗標會回傳錯誤
func TestDirectSendCmdMissingRequiredFlags(t *testing.T) {
	viper.Reset()
	root := &cobra.Command{Use: "hermes"}
	root.AddCommand(directSendMailCmd)
	// 重新綁定旗標到 viper，避免 Reset 後遺失設定
	viper.BindPFlag("host", directSendMailCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("port", directSendMailCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("from", directSendMailCmd.PersistentFlags().Lookup("from"))
	viper.BindPFlag("to", directSendMailCmd.PersistentFlags().Lookup("to"))
	viper.BindPFlag("subject", directSendMailCmd.PersistentFlags().Lookup("subject"))
	viper.BindPFlag("contents", directSendMailCmd.PersistentFlags().Lookup("contents"))
	viper.BindPFlag("cc", directSendMailCmd.PersistentFlags().Lookup("cc"))

	root.SetArgs([]string{"directSendMail"})
	_, err := root.ExecuteC()
	if err == nil {
		t.Fatalf("預期缺少旗標時應回傳錯誤")
	}
}

// TestDirectSendCmdRunCalled 確認提供旗標時 Run 會執行
func TestDirectSendCmdRunCalled(t *testing.T) {
	viper.Reset()
	root := &cobra.Command{Use: "hermes"}
	root.AddCommand(directSendMailCmd)
	// 重新綁定旗標到 viper，避免 Reset 後遺失設定
	viper.BindPFlag("host", directSendMailCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("port", directSendMailCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("from", directSendMailCmd.PersistentFlags().Lookup("from"))
	viper.BindPFlag("to", directSendMailCmd.PersistentFlags().Lookup("to"))
	viper.BindPFlag("subject", directSendMailCmd.PersistentFlags().Lookup("subject"))
	viper.BindPFlag("contents", directSendMailCmd.PersistentFlags().Lookup("contents"))
	viper.BindPFlag("cc", directSendMailCmd.PersistentFlags().Lookup("cc"))

	called := false
	originalRun := directSendMailCmd.Run
	directSendMailCmd.Run = func(cmd *cobra.Command, args []string) {
		sendmail.DirectSendMail(func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			called = true
			return nil
		})
	}
	defer func() { directSendMailCmd.Run = originalRun }()

	root.SetArgs([]string{
		"directSendMail",
		"--host", "smtp.example.com",
		"--port", "25",
		"--from", "from@example.com",
		"--to", "to@example.com",
		"--subject", "hi",
		"--contents", "body",
	})
	if _, err := root.ExecuteC(); err != nil {
		t.Fatalf("執行命令應無錯誤: %v", err)
	}
	if !called {
		t.Fatalf("預期郵件發送函式被呼叫")
	}
}
