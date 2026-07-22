package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/wtg42/hermes/mail"
	"github.com/wtg42/hermes/sendmail"
)

type historyRecordingMailer struct {
	calls   int
	compose mail.MailCompose
	err     error
}

func (m *historyRecordingMailer) Send(compose mail.MailCompose) error {
	m.calls++
	m.compose = compose
	return m.err
}

func executeHistoryCommand(t *testing.T, path func() (string, error), send structuredSendFunc, args ...string) (string, *cobra.Command, error) {
	t.Helper()
	root := &cobra.Command{Use: "hermes", SilenceErrors: true, SilenceUsage: true}
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.AddCommand(newHistoryCmd(path, send))
	root.SetArgs(append([]string{"history"}, args...))
	command, err := root.ExecuteC()
	return output.String(), command, err
}

func TestHistoryListCommand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	for index := range 12 {
		record := cmdHistoryRecord(string(rune('a'+index)), time.Date(2026, 7, 22, index, 0, 0, 0, time.UTC))
		if err := sendmail.AppendHistory(path, record); err != nil {
			t.Fatal(err)
		}
	}

	output, _, err := executeHistoryCommand(t, func() (string, error) { return path, nil }, nil, "list")
	if err != nil {
		t.Fatalf("history list error = %v", err)
	}
	var summaries []sendmail.HistorySummary
	if err := json.Unmarshal([]byte(output), &summaries); err != nil {
		t.Fatalf("history list output = %q: %v", output, err)
	}
	if len(summaries) != 10 || summaries[0].ID != "l" || summaries[9].ID != "c" {
		t.Fatalf("summaries = %+v", summaries)
	}
	for _, hidden := range []string{"body", "bcc", "attachments"} {
		if strings.Contains(output, `"`+hidden+`"`) {
			t.Fatalf("list output leaks %s: %s", hidden, output)
		}
	}

	output, _, err = executeHistoryCommand(t, func() (string, error) { return path, nil }, nil, "list", "--limit", "3")
	if err != nil || json.Unmarshal([]byte(output), &summaries) != nil || len(summaries) != 3 {
		t.Fatalf("limited list = %q, %v", output, err)
	}
	if _, _, err := executeHistoryCommand(t, func() (string, error) { return path, nil }, nil, "list", "--limit", "0"); err == nil {
		t.Fatal("history list accepted zero limit")
	}
}

func TestHistoryListEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.jsonl")
	output, _, err := executeHistoryCommand(t, func() (string, error) { return path, nil }, nil, "list")
	if err != nil || strings.TrimSpace(output) != "[]" {
		t.Fatalf("empty history list = %q, %v", output, err)
	}
}

func TestHistoryShowCommand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	record := cmdHistoryRecord("record-id", time.Now().UTC())
	record.Request.Body = "中文\nEnglish"
	if err := sendmail.AppendHistory(path, record); err != nil {
		t.Fatal(err)
	}

	output, _, err := executeHistoryCommand(t, func() (string, error) { return path, nil }, nil, "show", record.ID)
	if err != nil {
		t.Fatalf("history show error = %v", err)
	}
	var got sendmail.HistoryRecord
	if err := json.Unmarshal([]byte(output), &got); err != nil || got.ID != record.ID || got.Request.Body != record.Request.Body {
		t.Fatalf("history show = %+v, %v", got, err)
	}
	if _, _, err := executeHistoryCommand(t, func() (string, error) { return path, nil }, nil, "show", "missing"); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("missing show error = %v", err)
	}
	if _, _, err := executeHistoryCommand(t, func() (string, error) { return path, nil }, nil, "show"); err == nil {
		t.Fatal("history show accepted missing id")
	}
}

func TestHistoryCommandsFailClosedOnInvalidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	if err := os.WriteFile(path, []byte(`{"version":99}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"list"}, {"show", "id"}, {"replay", "id"}} {
		sendCalls := 0
		_, _, err := executeHistoryCommand(t, func() (string, error) { return path, nil }, func(sendmail.StructuredSendOptions) error {
			sendCalls++
			return nil
		}, args...)
		if err == nil || sendCalls != 0 {
			t.Fatalf("args %v: error = %v, send calls = %d", args, err, sendCalls)
		}
	}
}

func TestHistoryHelpHasNoSideEffects(t *testing.T) {
	pathCalls := 0
	sendCalls := 0
	for _, args := range [][]string{{"--help"}, {"list", "--help"}, {"show", "--help"}, {"replay", "--help"}} {
		_, _, err := executeHistoryCommand(t, func() (string, error) {
			pathCalls++
			return "", errors.New("must not resolve")
		}, func(sendmail.StructuredSendOptions) error {
			sendCalls++
			return nil
		}, args...)
		if err != nil {
			t.Fatalf("help %v error = %v", args, err)
		}
	}
	if pathCalls != 0 || sendCalls != 0 {
		t.Fatalf("help side effects: path %d send %d", pathCalls, sendCalls)
	}
}

func TestHistoryReplayCommand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	source := cmdHistoryRecord("source", time.Now().UTC())
	if err := sendmail.AppendHistory(path, source); err != nil {
		t.Fatal(err)
	}

	mailer := &historyRecordingMailer{}
	sender := sendmail.NewStructuredSender(func(plan sendmail.StructuredSendPlan) error { return mailer.Send(plan.Compose) })
	send := func(options sendmail.StructuredSendOptions) error {
		return sendmail.SendStructuredWithHistory(sender, options, path)
	}
	if _, _, err := executeHistoryCommand(t, func() (string, error) { return path, nil }, send, "replay", source.ID); err != nil {
		t.Fatalf("history replay error = %v", err)
	}
	if mailer.calls != 1 || mailer.compose.Subject != source.Request.Subject || mailer.compose.Body != source.Request.Body {
		t.Fatalf("mailer = calls %d compose %+v", mailer.calls, mailer.compose)
	}
	records, err := sendmail.ReadHistory(path)
	if err != nil || len(records) != 2 {
		t.Fatalf("replay history = %+v, %v", records, err)
	}

	if _, _, err := executeHistoryCommand(t, func() (string, error) { return path, nil }, send, "replay", source.ID, "--no-history"); err != nil {
		t.Fatalf("no-history replay error = %v", err)
	}
	records, _ = sendmail.ReadHistory(path)
	if mailer.calls != 2 || len(records) != 2 {
		t.Fatalf("no-history replay calls = %d records = %d", mailer.calls, len(records))
	}
}

func TestHistoryReplayRequiresFreshSafetyConfirmation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	record := cmdHistoryRecord("external", time.Now().UTC())
	record.Request.From = "sender@example.com"
	record.Request.To = []string{"to@example.com"}
	record.Request.CC = []string{"cc@example.net"}
	record.Request.BCC = []string{"bcc@example.org"}
	record.Request.Port = "1025"
	if err := sendmail.AppendHistory(path, record); err != nil {
		t.Fatal(err)
	}

	mailer := &historyRecordingMailer{}
	sender := sendmail.NewStructuredSender(func(plan sendmail.StructuredSendPlan) error { return mailer.Send(plan.Compose) })
	send := func(options sendmail.StructuredSendOptions) error {
		return sendmail.SendStructuredWithHistory(sender, options, path)
	}
	_, _, err := executeHistoryCommand(t, func() (string, error) { return path, nil }, send, "replay", record.ID, "--no-history")
	if err == nil || !strings.Contains(err.Error(), "sender@example.com") || !strings.Contains(err.Error(), "to@example.com") || !strings.Contains(err.Error(), "1025") || mailer.calls != 0 {
		t.Fatalf("unconfirmed replay error = %v, calls = %d", err, mailer.calls)
	}
	if _, _, err := executeHistoryCommand(t, func() (string, error) { return path, nil }, send, "replay", record.ID, "--confirm-outside-whitelist", "--no-history"); err != nil || mailer.calls != 1 {
		t.Fatalf("confirmed replay error = %v, calls = %d", err, mailer.calls)
	}
}

func TestHistoryReplayMissingAttachmentFailsBeforeMailer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	record := cmdHistoryRecord("attachment", time.Now().UTC())
	record.Request.Attachments = []string{filepath.Join(t.TempDir(), "missing.txt")}
	if err := sendmail.AppendHistory(path, record); err != nil {
		t.Fatal(err)
	}
	mailer := &historyRecordingMailer{}
	sender := sendmail.NewStructuredSender(func(plan sendmail.StructuredSendPlan) error { return mailer.Send(plan.Compose) })
	send := func(options sendmail.StructuredSendOptions) error {
		return sendmail.SendStructuredWithHistory(sender, options, path)
	}
	_, _, err := executeHistoryCommand(t, func() (string, error) { return path, nil }, send, "replay", record.ID)
	if err == nil || mailer.calls != 0 {
		t.Fatalf("missing attachment replay error = %v, calls = %d", err, mailer.calls)
	}
	records, _ := sendmail.ReadHistory(path)
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1", len(records))
	}
}

func TestRootRegistersHistoryCommand(t *testing.T) {
	command, _, err := rootCmd.Find([]string{"history"})
	if err != nil || command == rootCmd || command.Name() != "history" {
		t.Fatalf("root history command not registered: command=%v error=%v", command, err)
	}
}

func cmdHistoryRecord(id string, createdAt time.Time) sendmail.HistoryRecord {
	return sendmail.HistoryRecord{
		Version:   sendmail.HistoryRecordVersion,
		ID:        id,
		CreatedAt: createdAt,
		Request: sendmail.RecordedSendRequest{
			Server:   "192.0.2.10",
			Port:     "25",
			From:     "weitingshih@rd01.softnext.com.tw",
			To:       []string{"adam@rd01.softnext.com.tw"},
			CC:       []string{"jllee@rd01.softnext.com.tw"},
			BCC:      []string{"audit@rd01.softnext.com.tw"},
			Subject:  "History 中文 📨",
			Body:     "body",
			TLSMode:  sendmail.TLSModeNone,
			AuthMode: sendmail.AuthModeNone,
		},
		Result: sendmail.RecordedSendResult{Success: true},
	}
}
