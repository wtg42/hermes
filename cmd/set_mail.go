package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var SenderEmail string
var receiverEmail string
var emailSubject string
var emailBody string

var setMailCmd = &cobra.Command{
	Use:   "set",
	Short: "set command is used to set the mail info like sender or receiver",
	Long:  `set command is used to set the mail info like sender or receiver e.g. set --from="sender@example.com"`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(args)        // [set]
		fmt.Println(SenderEmail) // weiting.shi@gmail.com

		SenderEmail := viper.GetString("from")
		fmt.Println(SenderEmail)
	},
}

// 初始化時候設定這個命令的 flag
func init() {
	// 使用 set 這個命令時可以用 '--from' flag 來設定發件人電子郵件地址
	setMailCmd.PersistentFlags().StringVar(
		&SenderEmail,
		"from",
		"",
		"設定發件人電子郵件地址 (例如: 'someone.you.love@example.com')",
	)
	setMailCmd.MarkPersistentFlagRequired("from")

	// 使用 '--to' flag 來設定收件人電子郵件地址
	setMailCmd.PersistentFlags().StringVar(
		&receiverEmail,
		"to",
		"",
		"設定收件人電子郵件地址 (例如: 'someone.you.hate@example.com')",
	)
	setMailCmd.MarkPersistentFlagRequired("to")

	// 使用 '--subject' flag 來設定郵件主題
	setMailCmd.PersistentFlags().StringVar(&emailSubject, "subject", "", "設定郵件主題")
	setMailCmd.MarkPersistentFlagRequired("subject")

	// 使用 '--body' flag 來設定郵件內容
	setMailCmd.PersistentFlags().StringVar(&emailBody, "body", "", "設定郵件內容")
	setMailCmd.MarkPersistentFlagRequired("body")

	// 將 flag 綁定到 viper 配置中 統一管理且方便在其他檔案使用
	viper.BindPFlag("from", setMailCmd.PersistentFlags().Lookup("from"))
	viper.BindPFlag("to", setMailCmd.PersistentFlags().Lookup("to"))
	viper.BindPFlag("subject", setMailCmd.PersistentFlags().Lookup("subject"))
	viper.BindPFlag("body", setMailCmd.PersistentFlags().Lookup("body"))
}

func Execute() {
	if err := setMailCmd.Execute(); err != nil {
		fmt.Println("Error::::", err)
		os.Exit(1)
	}
}
