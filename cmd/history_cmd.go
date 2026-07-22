package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wtg42/hermes/sendmail"
)

type historyPathFunc func() (string, error)

var historyCmd = newHistoryCmd(sendmail.DefaultHistoryPath, runStructuredSend)

func newHistoryCmd(resolvePath historyPathFunc, send structuredSendFunc) *cobra.Command {
	history := &cobra.Command{
		Use:   "history",
		Short: "Inspect or replay structured send history.",
		Args:  cobra.NoArgs,
	}

	var limit int
	list := &cobra.Command{
		Use:   "list",
		Short: "List recent structured send history as JSON.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit < 1 {
				return fmt.Errorf("limit must be a positive integer")
			}
			path, err := resolvePath()
			if err != nil {
				return err
			}
			records, err := sendmail.ReadHistory(path)
			if err != nil {
				return err
			}
			recent := sendmail.RecentHistory(records, limit)
			summaries := make([]sendmail.HistorySummary, 0, len(recent))
			for _, record := range recent {
				summaries = append(summaries, sendmail.SummarizeHistory(record))
			}
			return writeHistoryJSON(cmd, summaries)
		},
	}
	list.Flags().IntVar(&limit, "limit", 10, "Maximum number of recent records")

	show := &cobra.Command{
		Use:   "show <id>",
		Short: "Show one complete structured send history record as JSON.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolvePath()
			if err != nil {
				return err
			}
			records, err := sendmail.ReadHistory(path)
			if err != nil {
				return err
			}
			record, found := sendmail.FindHistory(records, args[0])
			if !found {
				return fmt.Errorf("history record %q not found", args[0])
			}
			return writeHistoryJSON(cmd, record)
		},
	}

	var confirmOutsideWhitelist bool
	var noHistory bool
	var authPasswordStdin bool
	replay := &cobra.Command{
		Use:   "replay <id>",
		Short: "Replay one record through the current structured send pipeline.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolvePath()
			if err != nil {
				return err
			}
			records, err := sendmail.ReadHistory(path)
			if err != nil {
				return err
			}
			record, found := sendmail.FindHistory(records, args[0])
			if !found {
				return fmt.Errorf("history record %q not found", args[0])
			}
			if send == nil {
				return fmt.Errorf("structured send function is required")
			}
			options := sendmail.ReplayOptions(record)
			options.ConfirmOutsideWhitelist = confirmOutsideWhitelist
			options.NoHistory = noHistory
			options.AuthPasswordStdin = authPasswordStdin
			if err := sendmail.ValidateStructuredTransportStatic(options); err != nil {
				return err
			}
			if options.AuthPasswordStdin {
				if err := sendmail.ValidateStructuredSendStatic(options); err != nil {
					return err
				}
				password, readErr := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
				if readErr != nil && len(password) == 0 {
					return fmt.Errorf("read SMTP password from stdin: %w", readErr)
				}
				options.AuthPassword = strings.TrimSuffix(strings.TrimSuffix(password, "\n"), "\r")
				if options.AuthPassword == "" {
					return fmt.Errorf("SMTP password from stdin must not be empty")
				}
			}
			return send(options)
		},
	}
	replay.Flags().BoolVar(&confirmOutsideWhitelist, "confirm-outside-whitelist", false, "Explicitly authorize this replay outside the current safe whitelist")
	replay.Flags().BoolVar(&noHistory, "no-history", false, "Do not append a new history record for this replay")
	replay.Flags().BoolVar(&authPasswordStdin, "auth-password-stdin", false, "Read one fresh SMTP password line from stdin")

	history.AddCommand(list, show, replay)
	return history
}

func writeHistoryJSON(cmd *cobra.Command, value any) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("encode history output: %w", err)
	}
	return nil
}
