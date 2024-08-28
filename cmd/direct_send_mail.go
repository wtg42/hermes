// CLI 部分處理參數到呼叫郵件發送中間層
package cmd

import (
	"hermes/sendmail"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var host string
var SenderEmail string
var receiverEmail string
var emailSubject string
var emailBody string

var directSendMailCmd = &cobra.Command{
	Use:   "directSendMail",
	Short: "directSendMail command is used to set the mail info like sender or receiver",
	Long:  `directSendMail command is used to set the mail info like sender or receiver e.g. set --from="sender@example.com"`,
	Run: func(cmd *cobra.Command, args []string) {
		sendmail.DirectSendMail()
	},
}

// 初始化時候設定這個命令的 flag
func init() {
	directSendMailCmd.PersistentFlags().StringVar(&host, "host", "", "")
	directSendMailCmd.MarkPersistentFlagRequired("host")

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

	// 將 flag 綁定到 viper 配置中 統一管理且方便在其他檔案使用
	viper.BindPFlag("host", directSendMailCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("from", directSendMailCmd.PersistentFlags().Lookup("from"))
	viper.BindPFlag("to", directSendMailCmd.PersistentFlags().Lookup("to"))
	viper.BindPFlag("subject", directSendMailCmd.PersistentFlags().Lookup("subject"))
	viper.BindPFlag("contents", directSendMailCmd.PersistentFlags().Lookup("contents"))
}
