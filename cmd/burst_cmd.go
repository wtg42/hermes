// 使用多進程快速發信
// 適合壓力測試跟快速測試用
package cmd

import (
	"fmt"
	"strconv"
	"time"

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
	var runID string
	var subjectPrefix string
	var bodyKB int
	var workers int
	var ratePerSecond int
	var confirmBurst bool
	var asynchronous bool
	var queueURL string
	var queue string

	cmd := &cobra.Command{
		Use:   "burst",
		Short: "Burst Mode.",
		Long:  `Send mail in a burst of speed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			quantityToInt, err := strconv.Atoi(viper.GetString("burst-quantity"))
			if err != nil {
				return fmt.Errorf("invalid quantity: %w", err)
			}

			options := sendmail.BurstOptions{
				Quantity:       quantityToInt,
				Host:           viper.GetString("burst-host"),
				Port:           viper.GetString("burst-port"),
				From:           viper.GetString("burst-from"),
				To:             viper.GetString("burst-to"),
				Domains:        viper.GetStringSlice("burst-domain"),
				AllowedDomains: viper.GetStringSlice("burst-allow-domain"),
				RunID:          viper.GetString("burst-run-id"),
				SubjectPrefix:  viper.GetString("burst-subject-prefix"),
				BodyKB:         viper.GetInt("burst-body-kb"),
				Workers:        viper.GetInt("burst-workers"),
				RatePerSecond:  viper.GetInt("burst-rate"),
				ConfirmBurst:   viper.GetBool("burst-confirm"),
			}
			if viper.GetBool("burst-async") {
				return sendmail.EnqueueBurstToRabbitMQ(cmd.Context(), options, sendmail.RabbitQueueConfig{
					URL:   viper.GetString("burst-queue-url"),
					Queue: viper.GetString("burst-queue"),
				})
			}
			options.Progress = func(progress sendmail.BurstProgress) {
				fmt.Fprintf(cmd.OutOrStdout(), "burst progress run=%s attempted=%d/%d succeeded=%d failed=%d\n", progress.RunID, progress.Attempted, progress.Requested, progress.Succeeded, progress.Failed)
			}
			result, executeErr := sendmail.ExecuteBurst(options)
			if result.RunID != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "burst summary run=%s requested=%d attempted=%d succeeded=%d failed=%d duration=%s\n", result.RunID, result.Requested, result.Attempted, result.Succeeded, result.Failed, result.Duration.Round(time.Millisecond))
			}
			return executeErr
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
	cmd.PersistentFlags().StringVar(&runID, "run-id", "", "Traceable run identifier (auto-generated when omitted)")
	cmd.PersistentFlags().StringVar(&subjectPrefix, "subject-prefix", "HERMES-BURST", "Subject prefix for every generated message")
	cmd.PersistentFlags().IntVar(&bodyKB, "body-kb", 0, "Plain-text body size in KiB; 0 uses a small diagnostic body")
	cmd.PersistentFlags().IntVar(&workers, "workers", 0, "Concurrent SMTP workers; 0 uses the available CPU count")
	cmd.PersistentFlags().IntVar(&ratePerSecond, "rate", 0, "Maximum messages started per second; 0 is unlimited")
	cmd.PersistentFlags().BoolVar(&confirmBurst, "confirm-burst", false, "Confirm a run that sends more than 1000 messages")
	cmd.PersistentFlags().BoolVar(&asynchronous, "async", false, "Enqueue messages to RabbitMQ instead of sending directly")
	cmd.PersistentFlags().StringVar(&queueURL, "queue-url", "amqp://guest:guest@localhost:5672/", "RabbitMQ URL used with --async")
	cmd.PersistentFlags().StringVar(&queue, "queue", sendmail.DefaultDeliveryQueue, "RabbitMQ queue used with --async")

	_ = viper.BindPFlag("burst-quantity", cmd.PersistentFlags().Lookup("quantity"))
	_ = viper.BindPFlag("burst-host", cmd.PersistentFlags().Lookup("host"))
	_ = viper.BindPFlag("burst-port", cmd.PersistentFlags().Lookup("port"))
	_ = viper.BindPFlag("burst-from", cmd.PersistentFlags().Lookup("from"))
	_ = viper.BindPFlag("burst-to", cmd.PersistentFlags().Lookup("to"))
	_ = viper.BindPFlag("burst-domain", cmd.PersistentFlags().Lookup("domain"))
	_ = viper.BindPFlag("burst-allow-domain", cmd.PersistentFlags().Lookup("allow-domain"))
	_ = viper.BindPFlag("burst-run-id", cmd.PersistentFlags().Lookup("run-id"))
	_ = viper.BindPFlag("burst-subject-prefix", cmd.PersistentFlags().Lookup("subject-prefix"))
	_ = viper.BindPFlag("burst-body-kb", cmd.PersistentFlags().Lookup("body-kb"))
	_ = viper.BindPFlag("burst-workers", cmd.PersistentFlags().Lookup("workers"))
	_ = viper.BindPFlag("burst-rate", cmd.PersistentFlags().Lookup("rate"))
	_ = viper.BindPFlag("burst-confirm", cmd.PersistentFlags().Lookup("confirm-burst"))
	_ = viper.BindPFlag("burst-async", cmd.PersistentFlags().Lookup("async"))
	_ = viper.BindPFlag("burst-queue-url", cmd.PersistentFlags().Lookup("queue-url"))
	_ = viper.BindPFlag("burst-queue", cmd.PersistentFlags().Lookup("queue"))

	return cmd
}
