// 使用多進程快速發信
// 適合壓力測試跟快速測試用
package cmd

import (
	"hermes/sendmail"
	"log"

	"github.com/spf13/cobra"
)

var burstModeCmd = &cobra.Command{
	Use:   "burst",
	Short: "Burst Mode.",
	Long:  `Send mail in a burst of speed.`,
	Run: func(cmd *cobra.Command, args []string) {
		sendmail.BurstModeSendMail()
		// 跑多進程的函數
		log.Println("Burst Mode.")
	},
}
