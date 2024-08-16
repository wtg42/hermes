package main

import (
	"fmt"
	"go-go-power-mail/cmd"

	"github.com/spf13/viper"
)

// 定義用戶可以執行的指令
const (
	cmdStartTUI       = "start-tui"
	cmdDirectSendMail = "directSendMail"
)

func main() {
	// fetch user cmd
	cmd.Execute()
	userInputCmd := viper.Get("userInputCmd")
	fmt.Println("==>", userInputCmd)

	// 依照選擇指令
	switch userInputCmd {
	case cmdStartTUI:
		// TODO: 實現 TUI 啟動邏輯
	case cmdDirectSendMail:
		// sendmail.DirectSendMail()
	default:
		fmt.Printf("我不知道你想要幹嘛?: %s\n", userInputCmd)
	}
}
