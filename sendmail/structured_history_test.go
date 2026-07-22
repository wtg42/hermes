package sendmail

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStructuredSenderExecuteResult(t *testing.T) {
	t.Run("preflight failure", func(t *testing.T) {
		mailer := &recordingMailer{}
		execution, err := testStructuredSender(mailer).Execute(StructuredSendOptions{})
		if err == nil || execution.MailerAttempted || mailer.calls != 0 {
			t.Fatalf("Execute() = %+v, %v; mailer calls %d", execution, err, mailer.calls)
		}
	})

	t.Run("success", func(t *testing.T) {
		mailer := &recordingMailer{}
		execution, err := testStructuredSender(mailer).Execute(historySafeOptions())
		if err != nil || !execution.MailerAttempted || execution.MailerError != nil || mailer.calls != 1 {
			t.Fatalf("Execute() = %+v, %v; mailer calls %d", execution, err, mailer.calls)
		}
		if execution.Plan.Compose.Subject != "subject" || execution.Plan.Compose.Body != "body" {
			t.Fatalf("resolved compose = %+v", execution.Plan.Compose)
		}
	})

	t.Run("mailer failure", func(t *testing.T) {
		mailerErr := errors.New("smtp unavailable")
		mailer := &recordingMailer{err: mailerErr}
		execution, err := testStructuredSender(mailer).Execute(historySafeOptions())
		if !errors.Is(err, mailerErr) || !execution.MailerAttempted || !errors.Is(execution.MailerError, mailerErr) || mailer.calls != 1 {
			t.Fatalf("Execute() = %+v, %v; mailer calls %d", execution, err, mailer.calls)
		}
	})
}

func TestSendStructuredWithHistory(t *testing.T) {
	t.Run("records success", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "history.jsonl")
		mailer := &recordingMailer{}
		if err := SendStructuredWithHistory(testStructuredSender(mailer), historySafeOptions(), path); err != nil {
			t.Fatalf("SendStructuredWithHistory() error = %v", err)
		}
		records, err := ReadHistory(path)
		if err != nil || len(records) != 1 || !records[0].Result.Success {
			t.Fatalf("history = %+v, %v", records, err)
		}
	})

	t.Run("records smtp failure", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "history.jsonl")
		mailerErr := errors.New("smtp unavailable")
		mailer := &recordingMailer{err: mailerErr}
		err := SendStructuredWithHistory(testStructuredSender(mailer), historySafeOptions(), path)
		if !errors.Is(err, mailerErr) || mailer.calls != 1 {
			t.Fatalf("error = %v, calls = %d", err, mailer.calls)
		}
		records, readErr := ReadHistory(path)
		if readErr != nil || len(records) != 1 || records[0].Result.Success || records[0].Result.Error != mailerErr.Error() {
			t.Fatalf("history = %+v, %v", records, readErr)
		}
	})

	t.Run("preflight failure is not recorded", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "history.jsonl")
		mailer := &recordingMailer{}
		if err := SendStructuredWithHistory(testStructuredSender(mailer), StructuredSendOptions{}, path); err == nil {
			t.Fatal("expected preflight error")
		}
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) || mailer.calls != 0 {
			t.Fatalf("history stat error = %v, mailer calls = %d", err, mailer.calls)
		}
	})

	t.Run("no history performs no history io", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "history.jsonl")
		mailer := &recordingMailer{}
		options := historySafeOptions()
		options.NoHistory = true
		if err := SendStructuredWithHistory(testStructuredSender(mailer), options, path); err != nil {
			t.Fatalf("SendStructuredWithHistory() error = %v", err)
		}
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) || mailer.calls != 1 {
			t.Fatalf("history stat error = %v, mailer calls = %d", err, mailer.calls)
		}
	})
}

func TestSendStructuredWithHistorySeparatesSendAndWriteErrors(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(blocked, "history.jsonl")

	t.Run("mail sent", func(t *testing.T) {
		mailer := &recordingMailer{}
		err := SendStructuredWithHistory(testStructuredSender(mailer), historySafeOptions(), path)
		if err == nil || !strings.Contains(err.Error(), "mail was sent") || !strings.Contains(err.Error(), "history") || mailer.calls != 1 {
			t.Fatalf("error = %v, calls = %d", err, mailer.calls)
		}
	})

	t.Run("mail and history failed", func(t *testing.T) {
		mailerErr := errors.New("smtp unavailable")
		mailer := &recordingMailer{err: mailerErr}
		err := SendStructuredWithHistory(testStructuredSender(mailer), historySafeOptions(), path)
		if !errors.Is(err, mailerErr) || !strings.Contains(err.Error(), "history") || mailer.calls != 1 {
			t.Fatalf("error = %v, calls = %d", err, mailer.calls)
		}
	})
}

func TestStructuredSenderSendRemainsCompatible(t *testing.T) {
	mailer := &recordingMailer{}
	if err := testStructuredSender(mailer).Send(historySafeOptions()); err != nil || mailer.calls != 1 {
		t.Fatalf("Send() error = %v, calls = %d", err, mailer.calls)
	}
}

func historySafeOptions() StructuredSendOptions {
	return StructuredSendOptions{
		Server:  "192.0.2.10",
		Port:    "25",
		From:    "weitingshih@rd01.softnext.com.tw",
		To:      []string{"adam@rd01.softnext.com.tw"},
		Subject: "subject",
		Body:    "body",
	}
}
