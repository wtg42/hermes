// CLI 部分處理參數到呼叫郵件發送中間層
package cmd

import (
	"net/smtp"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wtg42/hermes/sendmail"
)

var directSendMailCmd = &cobra.Command{
	Use:   "directSendMail",
	Short: `"directSendMail" is a CLI command that quickly sends an email.`,
	Long:  `"directSendMail" is a CLI command that lets you send an email directly without using a TUI.`,
	Run: func(cmd *cobra.Command, args []string) {
		sendmail.DirectSendMail(smtp.SendMail)
	},
}

// 初始化時候設定這個命令的 flag
func init() {
	var host string
	var port string
	var SenderEmail string
	var receiverEmail string
	var ccEmail string
	var emailSubject string
	var emailBody string

	directSendMailCmd.PersistentFlags().StringVar(&host, "host", "", "MTA 主機名稱 (例如: 'smtp.gmail.com')")
	directSendMailCmd.MarkPersistentFlagRequired("host")

	directSendMailCmd.PersistentFlags().StringVar(&port, "port", "", "Port number (例如: '25')")
	directSendMailCmd.MarkPersistentFlagRequired("port")

	// 使用 directSendMail 這個命令時可以用 '--from' flag 來設定發件人電子郵件地址
	directSendMailCmd.PersistentFlags().StringVar(
		&SenderEmail,
		"from",
		"",
		"設定發件人電子郵件地址 (例如: 'someone.you.love@example.com')",
	)
	directSendMailCmd.MarkPersistentFlagRequired("from")

	// 使用 '--to' flag 來設定收件人電子郵件地址
	directSendMailCmd.PersistentFlags().StringVar(
		&receiverEmail,
		"to",
		"",
		"設定收件人電子郵件地址 (例如: 'someone.you.hate@example.com')",
	)
	directSendMailCmd.MarkPersistentFlagRequired("to")

	// 使用 '--subject' flag 來設定郵件主題
	directSendMailCmd.PersistentFlags().StringVar(&emailSubject, "subject", "", "設定郵件主題")
	directSendMailCmd.MarkPersistentFlagRequired("subject")

	// 使用 '--contents' flag 來設定郵件內容
	directSendMailCmd.PersistentFlags().StringVar(&emailBody, "contents", "", "設定郵件內容")
	directSendMailCmd.MarkPersistentFlagRequired("contents")

	// 使用 '--cc' flag 來設定副本電子郵件地址
	directSendMailCmd.PersistentFlags().StringVar(&ccEmail, "cc", "", "設定副本電子郵件地址 (可多個，以逗號分隔)")

	// 將 flag 綁定到 viper 配置中 統一管理且方便在其他檔案使用
	viper.BindPFlag("host", directSendMailCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("port", directSendMailCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("from", directSendMailCmd.PersistentFlags().Lookup("from"))
	viper.BindPFlag("to", directSendMailCmd.PersistentFlags().Lookup("to"))
	viper.BindPFlag("subject", directSendMailCmd.PersistentFlags().Lookup("subject"))
	viper.BindPFlag("contents", directSendMailCmd.PersistentFlags().Lookup("contents"))
	viper.BindPFlag("cc", directSendMailCmd.PersistentFlags().Lookup("cc"))
}
