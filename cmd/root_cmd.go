// 寫好的 command 以照需求放到 init() 做擴充
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "hermes",
	Short: "A command-line SMTP tool.",
	Long:  `A command-line tool for sending emails via SMTP.`,
}

func init() {
	rootCmd.AddCommand(directSendMailCmd)
	rootCmd.AddCommand(startTUICmd)
	rootCmd.AddCommand(burstModeCmd)
}

func Execute() {
	cmd, err := rootCmd.ExecuteC()
	if err != nil {
		fmt.Printf("Opps!: %v\n", err)
		os.Exit(1)
	}

	// 儲存用戶輸入的命令 後面也許用不到
	viper.Set("userInputCmd", cmd.Name())
}
