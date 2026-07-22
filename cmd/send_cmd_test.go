package cmd

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/wtg42/hermes/sendmail"
)

func executeSendCommand(t *testing.T, send func(sendmail.StructuredSendOptions) error, args ...string) (*cobra.Command, error) {
	t.Helper()
	root := &cobra.Command{Use: "hermes", SilenceErrors: true, SilenceUsage: true}
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.AddCommand(newSendCmd(send))
	root.SetArgs(append([]string{"send"}, args...))
	return root.ExecuteC()
}

func TestSendCmdCollectsMinimalOptions(t *testing.T) {
	var got sendmail.StructuredSendOptions
	calls := 0
	_, err := executeSendCommand(t, func(options sendmail.StructuredSendOptions) error {
		calls++
		got = options
		return nil
	}, "--server", "192.0.2.10")
	if err != nil {
		t.Fatalf("ExecuteC() error = %v", err)
	}
	if calls != 1 || got.Server != "192.0.2.10" {
		t.Fatalf("calls=%d options=%+v", calls, got)
	}
}

func TestSendCmdCollectsAllOptionsAndPreservesOrder(t *testing.T) {
	var got sendmail.StructuredSendOptions
	_, err := executeSendCommand(t, func(options sendmail.StructuredSendOptions) error {
		got = options
		return nil
	},
		"--server", "192.0.2.10",
		"--port", "1025",
		"--from", "sender@example.com",
		"--to", "to1@example.com,to2@example.com", "--to", "to3@example.com",
		"--cc", "cc1@example.com", "--cc", "cc2@example.com,cc3@example.com",
		"--bcc", "bcc1@example.com,bcc2@example.com",
		"--subject", "subject",
		"--body", "body",
		"--attach", "first.txt,second.txt", "--attach", "third.txt",
		"--confirm-outside-whitelist",
		"--no-history",
	)
	if err != nil {
		t.Fatalf("ExecuteC() error = %v", err)
	}

	want := sendmail.StructuredSendOptions{
		Server:                  "192.0.2.10",
		Port:                    "1025",
		From:                    "sender@example.com",
		To:                      []string{"to1@example.com", "to2@example.com", "to3@example.com"},
		CC:                      []string{"cc1@example.com", "cc2@example.com", "cc3@example.com"},
		BCC:                     []string{"bcc1@example.com", "bcc2@example.com"},
		Subject:                 "subject",
		Body:                    "body",
		Attachments:             []string{"first.txt", "second.txt", "third.txt"},
		ConfirmOutsideWhitelist: true,
		NoHistory:               true,
		TLSMode:                 sendmail.TLSModeNone,
		AuthMode:                sendmail.AuthModeNone,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("options = %#v, want %#v", got, want)
	}
}

func TestSendCmdReturnsServiceError(t *testing.T) {
	wantErr := errors.New("send failed")
	_, err := executeSendCommand(t, func(sendmail.StructuredSendOptions) error { return wantErr }, "--server", "192.0.2.10")
	if !errors.Is(err, wantErr) {
		t.Fatalf("ExecuteC() error = %v, want %v", err, wantErr)
	}
}

func TestSendCmdMissingServerDoesNotCallService(t *testing.T) {
	calls := 0
	_, err := executeSendCommand(t, func(sendmail.StructuredSendOptions) error {
		calls++
		return nil
	})
	if err == nil || calls != 0 {
		t.Fatalf("error=%v calls=%d", err, calls)
	}
}

func TestSendCmdHelpDoesNotCallService(t *testing.T) {
	calls := 0
	command, err := executeSendCommand(t, func(sendmail.StructuredSendOptions) error {
		calls++
		return nil
	}, "--help")
	if err != nil || calls != 0 || command.Name() != "send" {
		t.Fatalf("command=%q error=%v calls=%d", command.Name(), err, calls)
	}
}

func TestRootRegistersSendCommand(t *testing.T) {
	command, _, err := rootCmd.Find([]string{"send"})
	if err != nil || command == rootCmd || command.Name() != "send" {
		t.Fatalf("root send command not registered: command=%v error=%v", command, err)
	}
}
