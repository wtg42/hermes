package sendmail

import (
	"errors"
	"net/smtp"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/wtg42/hermes/mail"
)

func TestLegacyMailCompose(t *testing.T) {
	tests := []struct {
		name   string
		fields map[string]any
		want   mail.MailCompose
	}{
		{
			name: "full configuration",
			fields: map[string]any{
				"host": "smtp.example.com", "port": "2525", "from": "from@example.com",
				"to": "first@example.com,second@example.com", "cc": "cc@example.com",
				"bcc": "hidden@example.com", "subject": "subject", "contents": "body",
				"attachment": "/tmp/report.txt",
			},
			want: mail.MailCompose{
				Host: "smtp.example.com", Port: "2525", From: "from@example.com",
				To: []string{"first@example.com", "second@example.com"}, CC: []string{"cc@example.com"},
				BCC: []string{"hidden@example.com"}, Subject: "subject", Body: "body",
				Attachment: "/tmp/report.txt",
			},
		},
		{
			name: "defaults and optional fields",
			fields: map[string]any{
				"host": "smtp.example.com", "from": "from@example.com", "to": "to@example.com",
				"subject": "subject", "contents": "body",
			},
			want: mail.MailCompose{
				Host: "smtp.example.com", Port: "25", From: "from@example.com",
				To: []string{"to@example.com"}, Subject: "subject", Body: "body",
			},
		},
		{
			name: "empty port uses default",
			fields: map[string]any{
				"host": "smtp.example.com", "port": "", "from": "from@example.com",
				"to": "to@example.com", "cc": "", "bcc": "", "subject": "subject",
				"contents": "body", "attachment": "",
			},
			want: mail.MailCompose{
				Host: "smtp.example.com", Port: "25", From: "from@example.com",
				To: []string{"to@example.com"}, Subject: "subject", Body: "body",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := legacyMailCompose(tt.fields)
			if err != nil {
				t.Fatalf("legacyMailCompose() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("legacyMailCompose() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestLegacyMailComposePreservesCommaSeparatedAddressData(t *testing.T) {
	compose, err := legacyMailCompose(map[string]any{
		"host": "smtp.example.com", "from": "from@example.com",
		"to": "first@example.com, second@example.com", "cc": "invalid@,cc@example.com",
		"bcc": "", "subject": "subject", "contents": "body",
	})
	if err != nil {
		t.Fatalf("legacyMailCompose() error = %v", err)
	}

	if want := []string{"first@example.com", " second@example.com"}; !reflect.DeepEqual(compose.To, want) {
		t.Fatalf("To = %#v, want %#v", compose.To, want)
	}
	if want := []string{"invalid@", "cc@example.com"}; !reflect.DeepEqual(compose.CC, want) {
		t.Fatalf("CC = %#v, want %#v", compose.CC, want)
	}
	if compose.BCC != nil {
		t.Fatalf("BCC = %#v, want nil", compose.BCC)
	}
}

func TestLegacyMailComposeRejectsMissingOrInvalidRequiredFields(t *testing.T) {
	valid := map[string]any{
		"host": "smtp.example.com", "from": "from@example.com", "to": "to@example.com",
		"subject": "subject", "contents": "body",
	}

	for _, field := range []string{"host", "from", "to", "subject", "contents"} {
		t.Run("missing "+field, func(t *testing.T) {
			fields := cloneLegacyFields(valid)
			delete(fields, field)
			_, err := legacyMailCompose(fields)
			if err == nil || !strings.Contains(err.Error(), field) {
				t.Fatalf("error = %v, want error identifying %q", err, field)
			}
		})

		t.Run("invalid "+field, func(t *testing.T) {
			fields := cloneLegacyFields(valid)
			fields[field] = 42
			_, err := legacyMailCompose(fields)
			if err == nil || !strings.Contains(err.Error(), field) {
				t.Fatalf("error = %v, want error identifying %q", err, field)
			}
		})
	}
}

func TestSendMailWithMultipartRejectsInvalidLegacyConfigBeforeSMTP(t *testing.T) {
	original := SendMail
	t.Cleanup(func() { SendMail = original })

	calls := 0
	SendMail = func(string, smtp.Auth, string, []string, []byte) error {
		calls++
		return nil
	}
	viper.Reset()
	viper.Set("invalidLegacyMail", map[string]any{
		"host": "smtp.example.com", "from": "from@example.com", "to": 42,
		"subject": "subject", "contents": "body",
	})

	ok, err := SendMailWithMultipart("invalidLegacyMail")
	if ok || err == nil || !strings.Contains(err.Error(), "to") {
		t.Fatalf("SendMailWithMultipart() = (%v, %v), want false and to error", ok, err)
	}
	if calls != 0 {
		t.Fatalf("SMTP calls = %d, want 0", calls)
	}
}

func TestSendMailWithMultipartCallsCommonPipelineOnce(t *testing.T) {
	for _, tt := range []struct {
		name    string
		smtpErr error
		wantOK  bool
	}{
		{name: "success", wantOK: true},
		{name: "failure", smtpErr: errors.New("smtp unavailable")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			original := SendMail
			t.Cleanup(func() { SendMail = original })

			calls := 0
			SendMail = func(string, smtp.Auth, string, []string, []byte) error {
				calls++
				return tt.smtpErr
			}
			viper.Reset()
			viper.Set("legacyMail", map[string]any{
				"host": "smtp.example.com", "from": "from@example.com", "to": "to@example.com",
				"subject": "subject", "contents": "body",
			})

			ok, err := SendMailWithMultipart("legacyMail")
			if ok != tt.wantOK || !errors.Is(err, tt.smtpErr) {
				t.Fatalf("SendMailWithMultipart() = (%v, %v), want (%v, %v)", ok, err, tt.wantOK, tt.smtpErr)
			}
			if calls != 1 {
				t.Fatalf("SMTP calls = %d, want 1", calls)
			}
		})
	}
}

func cloneLegacyFields(fields map[string]any) map[string]any {
	cloned := make(map[string]any, len(fields))
	for key, value := range fields {
		cloned[key] = value
	}
	return cloned
}
