package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wtg42/hermes/sendmail"
)

type structuredSendFunc func(sendmail.StructuredSendOptions) error

func runStructuredSend(options sendmail.StructuredSendOptions) error {
	mailer := sendmail.NewSMTPMailer()
	sender := sendmail.NewStructuredSender(func(plan sendmail.StructuredSendPlan) error {
		return mailer.SendWithTransport(plan.Compose, plan.Transport)
	})
	if options.NoHistory {
		return sender.Send(options)
	}
	historyPath, err := sendmail.DefaultHistoryPath()
	if err != nil {
		return err
	}
	return sendmail.SendStructuredWithHistory(sender, options, historyPath)
}

var sendCmd = newSendCmd(runStructuredSend)

func newSendCmd(send structuredSendFunc) *cobra.Command {
	var options sendmail.StructuredSendOptions

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send one structured SMTP test message.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := sendmail.ValidateStructuredTransportStatic(options); err != nil {
				return err
			}
			if options.AuthPasswordStdin {
				if err := sendmail.ValidateStructuredSendStatic(options); err != nil {
					return err
				}
				password, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
				if err != nil && len(password) == 0 {
					return fmt.Errorf("read SMTP password from stdin: %w", err)
				}
				options.AuthPassword = strings.TrimSuffix(strings.TrimSuffix(password, "\n"), "\r")
				if options.AuthPassword == "" {
					return fmt.Errorf("SMTP password from stdin must not be empty")
				}
			}
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
	flags.BoolVar(&options.NoHistory, "no-history", false, "Do not save this send in local history")
	flags.StringVar(&options.TLSMode, "tls-mode", sendmail.TLSModeNone, "TLS policy: none or required")
	flags.StringVar(&options.TLSServerName, "tls-server-name", "", "Server identity used for TLS certificate verification")
	flags.StringVar(&options.AuthMode, "auth-mode", sendmail.AuthModeNone, "SMTP authentication: none or plain")
	flags.StringVar(&options.AuthUsername, "auth-username", "", "SMTP authentication username")
	flags.BoolVar(&options.AuthPasswordStdin, "auth-password-stdin", false, "Read one SMTP password line from stdin")
	_ = cmd.MarkFlagRequired("server")

	return cmd
}
