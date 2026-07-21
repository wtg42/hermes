package sendmail

import (
	"reflect"
	"strings"
	"testing"
)

func TestAssessSingleSendSafety(t *testing.T) {
	tests := []struct {
		name        string
		input       SingleSendSafetyInput
		wantReasons []string
	}{
		{
			name: "safe",
			input: SingleSendSafetyInput{
				From: "JLLEE@RD01.SOFTNEXT.COM.TW",
				To:   []string{"to@RD01.SOFTNEXT.COM.TW"},
				CC:   []string{"cc@rd01.softnext.com.tw"},
				BCC:  []string{"bcc@rd01.softnext.com.tw"},
				Port: "25",
			},
		},
		{
			name: "external sender",
			input: SingleSendSafetyInput{
				From: "sender@rd01.softnext.com.tw",
				To:   []string{"to@rd01.softnext.com.tw"},
				Port: "25",
			},
			wantReasons: []string{"sender sender@rd01.softnext.com.tw is outside the sender whitelist"},
		},
		{
			name: "subdomain is external",
			input: SingleSendSafetyInput{
				From: safeSingleSenders[0],
				To:   []string{"to@mail.rd01.softnext.com.tw"},
				Port: "25",
			},
			wantReasons: []string{"recipient to@mail.rd01.softnext.com.tw is outside rd01.softnext.com.tw"},
		},
		{
			name: "stable aggregate order",
			input: SingleSendSafetyInput{
				From: "sender@example.com",
				To:   []string{"to1@example.com", "to2@example.com"},
				CC:   []string{"cc@example.net"},
				BCC:  []string{"bcc@example.org"},
				Port: "1025",
			},
			wantReasons: []string{
				"sender sender@example.com is outside the sender whitelist",
				"recipient to1@example.com is outside rd01.softnext.com.tw",
				"recipient to2@example.com is outside rd01.softnext.com.tw",
				"recipient cc@example.net is outside rd01.softnext.com.tw",
				"recipient bcc@example.org is outside rd01.softnext.com.tw",
				"port 1025 is outside the safe port 25",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessment, err := AssessSingleSendSafety(tt.input)
			if err != nil {
				t.Fatalf("AssessSingleSendSafety() error = %v", err)
			}
			if !reflect.DeepEqual(assessment.Reasons, tt.wantReasons) {
				t.Fatalf("Reasons = %#v, want %#v", assessment.Reasons, tt.wantReasons)
			}
			if assessment.RequiresConfirmation() != (len(tt.wantReasons) > 0) {
				t.Fatalf("RequiresConfirmation() = %v", assessment.RequiresConfirmation())
			}
		})
	}
}

func TestAssessSingleSendSafetyRejectsInvalidInput(t *testing.T) {
	valid := SingleSendSafetyInput{
		From: safeSingleSenders[0],
		To:   []string{"to@rd01.softnext.com.tw"},
		Port: "25",
	}
	tests := []struct {
		name    string
		mutate  func(*SingleSendSafetyInput)
		wantErr string
	}{
		{name: "invalid from", mutate: func(input *SingleSendSafetyInput) { input.From = "bad" }, wantErr: "from"},
		{name: "empty to", mutate: func(input *SingleSendSafetyInput) { input.To = nil }, wantErr: "to"},
		{name: "invalid to", mutate: func(input *SingleSendSafetyInput) { input.To = []string{"bad"} }, wantErr: "to"},
		{name: "invalid cc", mutate: func(input *SingleSendSafetyInput) { input.CC = []string{"bad"} }, wantErr: "cc"},
		{name: "invalid bcc", mutate: func(input *SingleSendSafetyInput) { input.BCC = []string{"bad"} }, wantErr: "bcc"},
		{name: "empty port", mutate: func(input *SingleSendSafetyInput) { input.Port = "" }, wantErr: "port"},
		{name: "zero port", mutate: func(input *SingleSendSafetyInput) { input.Port = "0" }, wantErr: "port"},
		{name: "large port", mutate: func(input *SingleSendSafetyInput) { input.Port = "65536" }, wantErr: "port"},
		{name: "named port", mutate: func(input *SingleSendSafetyInput) { input.Port = "smtp" }, wantErr: "port"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := valid
			tt.mutate(&input)
			assessment, err := AssessSingleSendSafety(input)
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), tt.wantErr) {
				t.Fatalf("error = %v, want field %q", err, tt.wantErr)
			}
			if assessment.RequiresConfirmation() || len(assessment.Reasons) != 0 {
				t.Fatalf("invalid input returned confirmable assessment: %+v", assessment)
			}
		})
	}
}

func TestStructuredSenderKeepsOrderedSharedSafetyReasons(t *testing.T) {
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

	message := err.Error()
	previous := -1
	for _, reasonFragment := range []string{
		"sender sender@example.com",
		"recipient to@example.com",
		"recipient cc@example.net",
		"recipient bcc@example.org",
		"port 1025",
	} {
		index := strings.Index(message, reasonFragment)
		if index <= previous {
			t.Fatalf("reason %q is missing or out of order in %q", reasonFragment, message)
		}
		previous = index
	}
	if mailer.calls != 0 {
		t.Fatalf("mailer calls = %d, want 0", mailer.calls)
	}
}
