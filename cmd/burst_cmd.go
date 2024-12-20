// 使用多進程快速發信
// 適合壓力測試跟快速測試用
package cmd

import (
	"log"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wtg42/hermes/sendmail"
)

var burstModeCmd = &cobra.Command{
	Use:   "burst",
	Short: "Burst Mode.",
	Long:  `Send mail in a burst of speed.`,
	Run: func(cmd *cobra.Command, args []string) {
		quantity := viper.GetString("burst-quantity")
		host := viper.GetString("burst-host")
		port := viper.GetString("burst-port")
		receiverDomain := strings.Split(viper.GetString("burst-domain"), ",")

		// 可接受用戶輸入想要發送的數量
		quantityToInt, err := strconv.ParseInt(quantity, 10, 64)
		if err != nil {
			log.Fatalf("Failed to convert quantity to int: %+v", err)
		}

		// 爆發模式
		sendmail.BurstModeSendMail(int(quantityToInt), host, port, receiverDomain)
	},
}

func init() {
	var quantity string
	var host string
	var port string

	burstModeCmd.PersistentFlags().StringVar(&quantity, "quantity", "", "The quantity of emails you want to send")
	burstModeCmd.MarkPersistentFlagRequired("quantity")

	burstModeCmd.PersistentFlags().StringVar(&host, "host", "", "MTA 主機名稱 (例如: 'smtp.gmail.com')")
	burstModeCmd.MarkPersistentFlagRequired("host")

	burstModeCmd.PersistentFlags().StringVar(&port, "port", "", "Port number (例如: '25')")
	burstModeCmd.MarkPersistentFlagRequired("port")

	viper.BindPFlag("burst-quantity", burstModeCmd.PersistentFlags().Lookup("quantity"))
	viper.BindPFlag("burst-host", burstModeCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("burst-port", burstModeCmd.PersistentFlags().Lookup("port"))
}
