// 使用多進程快速發信
// 適合壓力測試跟快速測試用
package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wtg42/hermes/sendmail"
)

var burstModeCmd = newBurstModeCmd()

func newBurstModeCmd() *cobra.Command {
	var quantity string
	var host string
	var port string
	var from string
	var to string
	var domains []string
	var allowedDomains []string

	cmd := &cobra.Command{
		Use:   "burst",
		Short: "Burst Mode.",
		Long:  `Send mail in a burst of speed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			quantityToInt, err := strconv.Atoi(viper.GetString("burst-quantity"))
			if err != nil {
				return fmt.Errorf("invalid quantity: %w", err)
			}

			return sendmail.BurstModeSendMail(sendmail.BurstOptions{
				Quantity:       quantityToInt,
				Host:           viper.GetString("burst-host"),
				Port:           viper.GetString("burst-port"),
				From:           viper.GetString("burst-from"),
				To:             viper.GetString("burst-to"),
				Domains:        viper.GetStringSlice("burst-domain"),
				AllowedDomains: viper.GetStringSlice("burst-allow-domain"),
			})
		},
	}

	cmd.PersistentFlags().StringVar(&quantity, "quantity", "", "The quantity of emails you want to send")
	_ = cmd.MarkPersistentFlagRequired("quantity")

	cmd.PersistentFlags().StringVar(&host, "host", "", "MTA 主機名稱 (例如: 'smtp.gmail.com')")
	_ = cmd.MarkPersistentFlagRequired("host")

	cmd.PersistentFlags().StringVar(&port, "port", "", "Port number (例如: '25')")
	_ = cmd.MarkPersistentFlagRequired("port")

	cmd.PersistentFlags().StringVar(&from, "from", "", "Fixed sender address; omitted means random")
	cmd.PersistentFlags().StringVar(&to, "to", "", "Fixed recipient address; omitted means random")
	cmd.PersistentFlags().StringSliceVar(&domains, "domain", nil, "Domains for random addresses (repeatable or comma-separated)")
	cmd.PersistentFlags().StringSliceVar(&allowedDomains, "allow-domain", nil, "Explicitly authorize an unlisted domain for this run (repeatable)")

	_ = viper.BindPFlag("burst-quantity", cmd.PersistentFlags().Lookup("quantity"))
	_ = viper.BindPFlag("burst-host", cmd.PersistentFlags().Lookup("host"))
	_ = viper.BindPFlag("burst-port", cmd.PersistentFlags().Lookup("port"))
	_ = viper.BindPFlag("burst-from", cmd.PersistentFlags().Lookup("from"))
	_ = viper.BindPFlag("burst-to", cmd.PersistentFlags().Lookup("to"))
	_ = viper.BindPFlag("burst-domain", cmd.PersistentFlags().Lookup("domain"))
	_ = viper.BindPFlag("burst-allow-domain", cmd.PersistentFlags().Lookup("allow-domain"))

	return cmd
}
