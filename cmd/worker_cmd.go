package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wtg42/hermes/sendmail"
)

func newWorkerCmd() *cobra.Command {
	var queueURL string
	var queue string
	var concurrency int
	cmd := &cobra.Command{
		Use:   "worker",
		Short: "Consume queued email jobs and deliver them through SMTP.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sendmail.RunRabbitMQWorker(cmd.Context(), sendmail.WorkerConfig{
				URL: queueURL, Queue: queue, Concurrency: concurrency,
			})
		},
	}
	cmd.Flags().StringVar(&queueURL, "queue-url", "amqp://guest:guest@localhost:5672/", "RabbitMQ URL")
	cmd.Flags().StringVar(&queue, "queue", sendmail.DefaultDeliveryQueue, "RabbitMQ queue")
	cmd.Flags().IntVar(&concurrency, "concurrency", 1, "Number of concurrent SMTP deliveries")
	return cmd
}
