package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "A command-line tool",
	Long:  `A command-line tool that can perform various operations.`,
}

func init() {
	rootCmd.AddCommand(directSendMailCmd)
	rootCmd.AddCommand(startTUICmd)
}

func Execute() {
	cmd, err := rootCmd.ExecuteC()
	if err != nil {
		fmt.Printf("我不知道你想要幹嘛?: %v\n", err)
		os.Exit(1)
	}

	// 儲存用戶輸入的命令 後面也許用不到
	viper.Set("userInputCmd", cmd.Name())
}
