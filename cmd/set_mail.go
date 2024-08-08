package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var senderEmail string
var receiverEmail string
var emailSubject string
var emailBody string

var setMailCmd = &cobra.Command{
	Use:   "set",
  Short: "set command is used to set the mail info like sender or receiver",
  Long: `set command is used to set the mail info like sender or receiver e.g. set --from="sender@example.com"`,
  Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Good good, let the hate flow through you")
		fmt.Println(args) // [set]
		fmt.Println(senderEmail) // weiting.shi@gmail.com
  },
}

// 初始化時候設定這個命令的 flag
func init() {
	// 使用 set 這個命令時可以用 '--from' flag 來設定發件人電子郵件地址
	setMailCmd.PersistentFlags().StringVar(&senderEmail, "from", "", "設定發件人電子郵件地址 (例如: 'someone.you.love@example.com')")

	// 使用 '--to' flag 來設定收件人電子郵件地址
	setMailCmd.PersistentFlags().StringVar(&receiverEmail, "to", "", "設定收件人電子郵件地址 (例如: 'someone.you.hate@example.com')")

	// 使用 '--subject' flag 來設定郵件主題
	setMailCmd.PersistentFlags().StringVar(&emailSubject, "subject", "", "設定郵件主題")

	// 使用 '--body' flag 來設定郵件內容
	setMailCmd.PersistentFlags().StringVar(&emailBody, "body", "", "設定郵件內容")
}

func Execute() {
	setMailCmd.Execute()
}