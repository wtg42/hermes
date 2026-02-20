// 寫好的 command 以照需求放到 init() 做擴充
package cmd

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wtg42/hermes/tui"
)

var rootCmd = &cobra.Command{
	Use:   "hermes",
	Short: "A command-line SMTP tool.",
	Long:  `A command-line tool for sending emails via SMTP.`,
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(tui.InitialMailFieldsModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatalf("發生錯誤：%v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(burstModeCmd)
	rootCmd.AddCommand(emlCmd)
}

// Execute 執行根命令並處理錯誤
//   - 會將用戶最後一次輸入的命令名稱保存到 viper
func Execute() {
	cmd, err := rootCmd.ExecuteC()
	if err != nil {
		fmt.Printf("Opps!: %v\n", err)
		os.Exit(1)
	}

	// 儲存用戶輸入的命令 後面也許用不到
	viper.Set("userInputCmd", cmd.Name())
}
