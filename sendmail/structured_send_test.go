package sendmail

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/wtg42/hermes/mail"
)

type recordingMailer struct {
	calls    int
	composes []mail.MailCompose
	err      error
}

func (m *recordingMailer) Send(compose mail.MailCompose) error {
	m.calls++
	m.composes = append(m.composes, compose)
	return m.err
}

func TestGenerateRandomContent(t *testing.T) {
	now := time.Date(2026, time.July, 21, 1, 2, 3, 0, time.UTC)
	subject, body, err := generateRandomContent(now, bytes.NewReader([]byte{1, 2, 3, 4, 0xde, 0xad, 0xbe, 0xef}))
	if err != nil {
		t.Fatalf("generateRandomContent() error = %v", err)
	}

	if !strings.ContainsAny(subject, "🚀📨🧪✨") {
		t.Fatalf("subject should contain emoji, got %q", subject)
	}
	if !strings.Contains(subject, "中文") || !strings.Contains(subject, "English") {
		t.Fatalf("subject should contain Chinese and English, got %q", subject)
	}
	if !strings.Contains(body, "中文") || !strings.Contains(body, "English") {
		t.Fatalf("body should contain Chinese and English, got %q", body)
	}
	if !strings.Contains(body, "2026-07-21 09:02:03 +08:00") {
		t.Fatalf("body should contain Taipei timestamp, got %q", body)
	}
	if !strings.Contains(body, "Trace-ID: deadbeef") {
		t.Fatalf("body should contain deterministic trace ID, got %q", body)
	}
}

func TestStructuredSenderOnlyFillsMissingContent(t *testing.T) {
	mailer := &recordingMailer{}
	sender := testStructuredSender(mailer)

	err := sender.Send(StructuredSendOptions{
		Server:  "192.0.2.10",
		Subject: "User subject",
		Body:    "User body exactly",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	compose := mailer.composes[0]
	if compose.Subject != "User subject" || compose.Body != "User body exactly" {
		t.Fatalf("user content was changed: subject=%q body=%q", compose.Subject, compose.Body)
	}

	mailer.composes = nil
	err = sender.Send(StructuredSendOptions{Server: "192.0.2.10", Subject: "Keep this"})
	if err != nil {
		t.Fatalf("Send() with missing body error = %v", err)
	}
	compose = mailer.composes[0]
	if compose.Subject != "Keep this" || !strings.Contains(compose.Body, "Trace-ID: deadbeef") {
		t.Fatalf("missing body was not filled independently: %+v", compose)
	}
}

func TestStructuredSenderSafeDefaults(t *testing.T) {
	for _, from := range append([]string{""}, safeSingleSenders...) {
		t.Run(from, func(t *testing.T) {
			mailer := &recordingMailer{}
			sender := testStructuredSender(mailer)
			if err := sender.Send(StructuredSendOptions{Server: "192.0.2.10", From: from}); err != nil {
				t.Fatalf("Send() error = %v", err)
			}

			compose := mailer.composes[0]
			wantFrom := from
			if wantFrom == "" {
				wantFrom = safeSingleSenders[0]
			}
			if compose.From != wantFrom || compose.Port != "25" {
				t.Fatalf("defaults = from %q port %q, want from %q port 25", compose.From, compose.Port, wantFrom)
			}
			if len(compose.To) != 1 || strings.EqualFold(compose.To[0], compose.From) {
				t.Fatalf("default To must differ from From: from=%q to=%v", compose.From, compose.To)
			}
		})
	}
}

func TestStructuredSenderRejectsInvalidServerEvenWithConfirmation(t *testing.T) {
	for _, server := range []string{"", "mail.example.com", "localhost", "::1", "192.0.2.999"} {
		t.Run(server, func(t *testing.T) {
			mailer := &recordingMailer{}
			err := testStructuredSender(mailer).Send(StructuredSendOptions{
				Server:                  server,
				ConfirmOutsideWhitelist: true,
			})
			if err == nil || mailer.calls != 0 {
				t.Fatalf("invalid server %q: error=%v calls=%d", server, err, mailer.calls)
			}
		})
	}
}

func TestStructuredSenderRejectsInvalidAddressesAndPorts(t *testing.T) {
	tests := []StructuredSendOptions{
		{Server: "192.0.2.10", From: "not-an-email", ConfirmOutsideWhitelist: true},
		{Server: "192.0.2.10", To: []string{"bad"}, ConfirmOutsideWhitelist: true},
		{Server: "192.0.2.10", CC: []string{"bad"}, ConfirmOutsideWhitelist: true},
		{Server: "192.0.2.10", BCC: []string{"bad"}, ConfirmOutsideWhitelist: true},
		{Server: "192.0.2.10", Port: "0", ConfirmOutsideWhitelist: true},
		{Server: "192.0.2.10", Port: "65536", ConfirmOutsideWhitelist: true},
		{Server: "192.0.2.10", Port: "smtp", ConfirmOutsideWhitelist: true},
	}
	for _, options := range tests {
		mailer := &recordingMailer{}
		if err := testStructuredSender(mailer).Send(options); err == nil || mailer.calls != 0 {
			t.Fatalf("options %+v: error=%v calls=%d", options, err, mailer.calls)
		}
	}
}

func TestStructuredSenderWhitelistPolicy(t *testing.T) {
	tests := []struct {
		name    string
		options StructuredSendOptions
		unsafe  bool
	}{
		{name: "safe sender and domain", options: StructuredSendOptions{From: safeSingleSenders[1], To: []string{"user@rd01.softnext.com.tw"}}},
		{name: "unsafe sender", options: StructuredSendOptions{From: "user@rd01.softnext.com.tw"}, unsafe: true},
		{name: "domain suffix", options: StructuredSendOptions{To: []string{"user@evilrd01.softnext.com.tw"}}, unsafe: true},
		{name: "subdomain", options: StructuredSendOptions{To: []string{"user@sub.rd01.softnext.com.tw"}}, unsafe: true},
		{name: "external to", options: StructuredSendOptions{To: []string{"user@gmail.com"}}, unsafe: true},
		{name: "external cc", options: StructuredSendOptions{CC: []string{"user@gmail.com"}}, unsafe: true},
		{name: "external bcc", options: StructuredSendOptions{BCC: []string{"user@gmail.com"}}, unsafe: true},
		{name: "non-safe port", options: StructuredSendOptions{Port: "1025"}, unsafe: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.options.Server = "192.0.2.10"
			mailer := &recordingMailer{}
			err := testStructuredSender(mailer).Send(tt.options)
			if tt.unsafe {
				if err == nil || !strings.Contains(err.Error(), "--confirm-outside-whitelist") || mailer.calls != 0 {
					t.Fatalf("unsafe request: error=%v calls=%d", err, mailer.calls)
				}
				tt.options.ConfirmOutsideWhitelist = true
				if err := testStructuredSender(mailer).Send(tt.options); err != nil {
					t.Fatalf("confirmed Send() error = %v", err)
				}
				return
			}
			if err != nil || mailer.calls != 1 {
				t.Fatalf("safe request: error=%v calls=%d", err, mailer.calls)
			}
		})
	}
}

func TestStructuredSenderAggregatesConfirmationReasons(t *testing.T) {
	mailer := &recordingMailer{}
	err := testStructuredSender(mailer).Send(StructuredSendOptions{
		Server: "192.0.2.10",
		Port:   "1025",
		From:   "sender@example.com",
		To:     []string{"to@example.com"},
		CC:     []string{"cc@example.net"},
		BCC:    []string{"bcc@example.org"},
	})
	if err == nil {
		t.Fatal("Send() expected confirmation error")
	}
	for _, want := range []string{"sender", "to@example.com", "cc@example.net", "bcc@example.org", "1025"} {
		if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(want)) {
			t.Errorf("error %q does not include %q", err, want)
		}
	}
}

func TestStructuredSenderPreflightHasZeroMailerSideEffects(t *testing.T) {
	missing := t.TempDir() + "/missing.txt"
	mailer := &recordingMailer{err: errors.New("must not be reached")}
	sender := testStructuredSender(mailer)

	for _, options := range []StructuredSendOptions{
		{Server: "hostname.invalid"},
		{Server: "192.0.2.10", To: []string{"outside@example.com"}},
		{Server: "192.0.2.10", Attachments: []string{missing}},
	} {
		if err := sender.Send(options); err == nil {
			t.Fatalf("Send(%+v) expected error", options)
		}
	}
	if mailer.calls != 0 {
		t.Fatalf("mailer calls = %d, want 0", mailer.calls)
	}
}

func testStructuredSender(mailer mail.Mailer) *StructuredSender {
	return &StructuredSender{
		mailer:  mailer,
		now:     func() time.Time { return time.Date(2026, time.July, 21, 1, 2, 3, 0, time.UTC) },
		entropy: bytes.NewReader(bytes.Repeat([]byte{1, 2, 3, 4, 0xde, 0xad, 0xbe, 0xef}, 32)),
	}
}
