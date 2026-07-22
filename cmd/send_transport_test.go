package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/wtg42/hermes/sendmail"
)

func TestSendCmdReadsPasswordAfterStaticPreflight(t *testing.T) {
	var got sendmail.StructuredSendOptions
	root := &cobra.Command{Use: "hermes", SilenceErrors: true, SilenceUsage: true}
	root.SetIn(strings.NewReader("one-shot-secret\n"))
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.AddCommand(newSendCmd(func(options sendmail.StructuredSendOptions) error { got = options; return nil }))
	root.SetArgs([]string{"send", "--server", "192.0.2.10", "--tls-mode", "required", "--tls-server-name", "smtp.test", "--auth-mode", "plain", "--auth-username", "user", "--auth-password-stdin"})
	if _, err := root.ExecuteC(); err != nil {
		t.Fatal(err)
	}
	if got.AuthPassword != "one-shot-secret" || got.AuthMode != sendmail.AuthModePlain {
		t.Fatalf("options = %+v", got)
	}
}

func TestSendCmdDoesNotReadPasswordOnStaticFailure(t *testing.T) {
	reader := &countingReader{Reader: strings.NewReader("must-not-read\n")}
	root := &cobra.Command{Use: "hermes", SilenceErrors: true, SilenceUsage: true}
	root.SetIn(reader)
	root.AddCommand(newSendCmd(func(sendmail.StructuredSendOptions) error { t.Fatal("send called"); return nil }))
	root.SetArgs([]string{"send", "--server", "invalid", "--tls-mode", "required", "--tls-server-name", "smtp.test", "--auth-mode", "plain", "--auth-username", "user", "--auth-password-stdin"})
	if _, err := root.ExecuteC(); err == nil || reader.reads != 0 {
		t.Fatalf("error=%v reads=%d", err, reader.reads)
	}
}

type countingReader struct {
	*strings.Reader
	reads int
}

func (reader *countingReader) Read(buffer []byte) (int, error) {
	reader.reads++
	return reader.Reader.Read(buffer)
}
