// 直接從 .eml 檔案載入並發送郵件
package cmd

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/wtg42/hermes/sendmail"
	"github.com/wtg42/hermes/tui"
)

var emlCmd = &cobra.Command{
	Use:   "eml",
	Short: "使用 .eml 檔案發送郵件",
	Long:  `從 .eml 檔案載入郵件內容並發送。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 建立 SMTP 郵件發送器
		mailer := sendmail.NewSMTPMailer()
		// 初始化 EML Model 並注入 mailer
		emlModel := tui.InitialEmlModel(mailer)
		initCmd := emlModel.Init()
		p := tea.NewProgram(emlModel, tea.WithAltScreen())
		if initCmd != nil {
			p.Send(initCmd())
		}
		if _, err := p.Run(); err != nil {
			log.Fatalf("發生錯誤：%v", err)
		}
	},
}
