package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wtg42/hermes/sendmail"
)

type structuredSendFunc func(sendmail.StructuredSendOptions) error

var sendCmd = newSendCmd(func(options sendmail.StructuredSendOptions) error {
	sender := sendmail.NewStructuredSender(sendmail.NewSMTPMailer())
	return sender.Send(options)
})

func newSendCmd(send structuredSendFunc) *cobra.Command {
	var options sendmail.StructuredSendOptions

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send one structured SMTP test message.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return send(options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.Server, "server", "", "SMTP server as an explicit dotted IPv4 address")
	flags.StringVar(&options.Port, "port", "", "SMTP port (default 25)")
	flags.StringVar(&options.From, "from", "", "Sender address (safe default when omitted)")
	flags.StringSliceVar(&options.To, "to", nil, "To addresses (repeatable or comma-separated)")
	flags.StringSliceVar(&options.CC, "cc", nil, "CC addresses (repeatable or comma-separated)")
	flags.StringSliceVar(&options.BCC, "bcc", nil, "BCC addresses (repeatable or comma-separated)")
	flags.StringVar(&options.Subject, "subject", "", "Message subject (random test subject when omitted)")
	flags.StringVar(&options.Body, "body", "", "Message body (random test body when omitted)")
	flags.StringSliceVar(&options.Attachments, "attach", nil, "Attachment paths (repeatable or comma-separated)")
	flags.BoolVar(&options.ConfirmOutsideWhitelist, "confirm-outside-whitelist", false, "Explicitly authorize recipients, sender, or port outside the safe whitelist")
	_ = cmd.MarkFlagRequired("server")

	return cmd
}
