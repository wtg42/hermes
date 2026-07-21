package tui

import (
	"net/smtp"
	"strings"
	"testing"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/viper"
	"github.com/wtg42/hermes/mail"
	"github.com/wtg42/hermes/sendmail"
)

type composeRecordingMailer struct {
	calls    int
	composes []mail.MailCompose
	err      error
}

func (m *composeRecordingMailer) Send(compose mail.MailCompose) error {
	m.calls++
	m.composes = append(m.composes, compose)
	return m.err
}

func TestComposeSafetyRoutesCtrlS(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*ComposeModel)
		wantCommand bool
		wantPending bool
		wantError   string
	}{
		{name: "safe message", wantCommand: true},
		{
			name:      "invalid address",
			mutate:    func(model *ComposeModel) { model.mailFields[1].SetValue("bad") },
			wantError: "to",
		},
		{
			name:        "external message",
			mutate:      makeComposeExternal,
			wantPending: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := &composeRecordingMailer{}
			model := newSafetyTestComposeModel(t, mailer)
			if tt.mutate != nil {
				tt.mutate(&model)
			}

			updated, cmd := updateComposeModel(t, model, ctrlKey('s'))
			if (cmd != nil) != tt.wantCommand {
				t.Fatalf("command present = %v, want %v", cmd != nil, tt.wantCommand)
			}
			if (updated.pendingConfirmation != nil) != tt.wantPending {
				t.Fatalf("pending = %v, want %v", updated.pendingConfirmation != nil, tt.wantPending)
			}
			if tt.wantError != "" && (updated.err == nil || !strings.Contains(strings.ToLower(updated.err.Error()), tt.wantError)) {
				t.Fatalf("error = %v, want field %q", updated.err, tt.wantError)
			}
			if mailer.calls != 0 {
				t.Fatalf("mailer calls before command execution = %d, want 0", mailer.calls)
			}

			if cmd != nil {
				cmd()
				if mailer.calls != 1 {
					t.Fatalf("mailer calls after command execution = %d, want 1", mailer.calls)
				}
			}
		})
	}
}

func TestComposeExternalConfirmationInput(t *testing.T) {
	tests := []struct {
		name         string
		confirmation string
		wantCommand  bool
	}{
		{name: "empty"},
		{name: "lowercase", confirmation: "send"},
		{name: "other", confirmation: "YES"},
		{name: "exact", confirmation: "SEND", wantCommand: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := &composeRecordingMailer{}
			model := newSafetyTestComposeModel(t, mailer)
			makeComposeExternal(&model)
			model, _ = updateComposeModel(t, model, ctrlKey('s'))
			if model.pendingConfirmation == nil {
				t.Fatal("expected pending confirmation")
			}

			for _, char := range tt.confirmation {
				model, _ = updateComposeModel(t, model, textKey(char))
			}
			updated, cmd := updateComposeModel(t, model, tea.KeyPressMsg{Code: tea.KeyEnter})
			if (cmd != nil) != tt.wantCommand {
				t.Fatalf("command present = %v, want %v", cmd != nil, tt.wantCommand)
			}
			if tt.wantCommand {
				if updated.pendingConfirmation != nil || !updated.sending {
					t.Fatalf("confirmed state = pending %v sending %v", updated.pendingConfirmation != nil, updated.sending)
				}
				cmd()
				if mailer.calls != 1 {
					t.Fatalf("mailer calls = %d, want 1", mailer.calls)
				}

				updated.sending = false
				reopened, reopenedCmd := updateComposeModel(t, updated, ctrlKey('s'))
				if reopenedCmd != nil || reopened.pendingConfirmation == nil {
					t.Fatal("authorization was reused instead of requesting confirmation again")
				}
				return
			}

			if updated.pendingConfirmation == nil || updated.pendingConfirmation.err == "" {
				t.Fatal("invalid confirmation should remain pending with an error")
			}
			if mailer.calls != 0 {
				t.Fatalf("mailer calls = %d, want 0", mailer.calls)
			}
		})
	}
}

func TestComposeExternalConfirmationEscPreservesMessage(t *testing.T) {
	mailer := &composeRecordingMailer{}
	model := newSafetyTestComposeModel(t, mailer)
	makeComposeExternal(&model)
	want := model.currentCompose()
	model, _ = updateComposeModel(t, model, ctrlKey('s'))

	updated, cmd := updateComposeModel(t, model, tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd != nil || updated.pendingConfirmation != nil {
		t.Fatalf("cancel result = command %v pending %v", cmd != nil, updated.pendingConfirmation != nil)
	}
	if !mailComposeSnapshotsEqual(updated.currentCompose(), want) {
		t.Fatalf("message changed after cancel: got %+v want %+v", updated.currentCompose(), want)
	}
	if mailer.calls != 0 {
		t.Fatalf("mailer calls = %d, want 0", mailer.calls)
	}
}

func TestMailComposeSnapshotsEqualProtectsEveryField(t *testing.T) {
	baseline := mail.MailCompose{
		From:        "sender@example.com",
		To:          []string{"to@example.com"},
		CC:          []string{"cc@example.com"},
		BCC:         []string{"bcc@example.com"},
		Subject:     "subject",
		Body:        "body",
		Attachment:  "/tmp/one.txt",
		Attachments: []string{"/tmp/one.txt", "/tmp/two.txt"},
		Host:        "smtp.example.com",
		Port:        "25",
	}
	mutations := []struct {
		name   string
		mutate func(*mail.MailCompose)
	}{
		{name: "from", mutate: func(compose *mail.MailCompose) { compose.From = "other@example.com" }},
		{name: "to", mutate: func(compose *mail.MailCompose) { compose.To = []string{"other@example.com"} }},
		{name: "cc", mutate: func(compose *mail.MailCompose) { compose.CC = []string{"other@example.com"} }},
		{name: "bcc", mutate: func(compose *mail.MailCompose) { compose.BCC = []string{"other@example.com"} }},
		{name: "subject", mutate: func(compose *mail.MailCompose) { compose.Subject = "other" }},
		{name: "body", mutate: func(compose *mail.MailCompose) { compose.Body = "other" }},
		{name: "attachment", mutate: func(compose *mail.MailCompose) { compose.Attachment = "/tmp/other.txt" }},
		{name: "attachments", mutate: func(compose *mail.MailCompose) { compose.Attachments = []string{"/tmp/other.txt"} }},
		{name: "host", mutate: func(compose *mail.MailCompose) { compose.Host = "other.example.com" }},
		{name: "port", mutate: func(compose *mail.MailCompose) { compose.Port = "1025" }},
	}

	if !mailComposeSnapshotsEqual(baseline, baseline) {
		t.Fatal("identical compose snapshots should match")
	}
	for _, tt := range mutations {
		t.Run(tt.name, func(t *testing.T) {
			changed := baseline
			tt.mutate(&changed)
			if mailComposeSnapshotsEqual(baseline, changed) {
				t.Fatalf("snapshot mutation %q was not detected", tt.name)
			}
		})
	}
}

func TestComposeConfirmationRejectsChangedSnapshot(t *testing.T) {
	mailer := &composeRecordingMailer{}
	model := newSafetyTestComposeModel(t, mailer)
	makeComposeExternal(&model)
	model, _ = updateComposeModel(t, model, ctrlKey('s'))
	model.mailFields[4].SetValue("changed after review")
	model.pendingConfirmation.input.SetValue("SEND")

	updated, cmd := updateComposeModel(t, model, tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil || updated.pendingConfirmation != nil {
		t.Fatalf("changed snapshot result = command %v pending %v", cmd != nil, updated.pendingConfirmation != nil)
	}
	if updated.err == nil || !strings.Contains(updated.err.Error(), "changed") {
		t.Fatalf("error = %v, want changed snapshot error", updated.err)
	}
	if mailer.calls != 0 {
		t.Fatalf("mailer calls = %d, want 0", mailer.calls)
	}
}

func TestComposeConfirmationIsolatesShortcuts(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyPressMsg
	}{
		{name: "send", key: ctrlKey('s')},
		{name: "attachment", key: ctrlKey('a')},
		{name: "html template", key: ctrlKey('h')},
		{name: "text template", key: ctrlKey('t')},
		{name: "eml template", key: ctrlKey('e')},
		{name: "ordinary text", key: textKey('X')},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := &composeRecordingMailer{}
			model := newSafetyTestComposeModel(t, mailer)
			model.activePanel = 1
			makeComposeExternal(&model)
			model, _ = updateComposeModel(t, model, ctrlKey('s'))
			before := model.currentCompose()

			updated, _ := updateComposeModel(t, model, tt.key)
			if !mailComposeSnapshotsEqual(updated.currentCompose(), before) {
				t.Fatalf("compose changed after confirmation key %q", tt.name)
			}
			if updated.showFilePicker {
				t.Fatalf("file picker opened after confirmation key %q", tt.name)
			}
			if updated.pendingConfirmation == nil {
				t.Fatalf("confirmation closed after key %q", tt.name)
			}
			if mailer.calls != 0 {
				t.Fatalf("mailer calls = %d, want 0", mailer.calls)
			}
		})
	}
}

func TestComposeConfirmationViewShowsReasonsAndInstructions(t *testing.T) {
	model := newSafetyTestComposeModel(t, &composeRecordingMailer{})
	makeComposeExternal(&model)
	model, _ = updateComposeModel(t, model, ctrlKey('s'))

	content := ansi.Strip(model.View().Content)
	for _, fragment := range []string{
		"sender sender@example.com",
		"recipient to@example.com",
		"recipient cc@example.net",
		"recipient bcc@example.org",
		"port 1025",
		"Type SEND to confirm",
		"Esc",
	} {
		if !strings.Contains(content, fragment) {
			t.Errorf("confirmation view missing %q", fragment)
		}
	}

	model, _ = updateComposeModel(t, model, tea.KeyPressMsg{Code: tea.KeyEnter})
	content = ansi.Strip(model.View().Content)
	if !strings.Contains(content, "Confirmation must be exactly SEND") {
		t.Fatal("confirmation input error is not visible")
	}
}

func TestComposeConfirmedMissingAttachmentDoesNotCallSMTP(t *testing.T) {
	original := sendmail.SendMail
	t.Cleanup(func() { sendmail.SendMail = original })
	smtpCalls := 0
	sendmail.SendMail = func(addr string, auth smtp.Auth, from string, to []string, message []byte) error {
		smtpCalls++
		return nil
	}

	model := newSafetyTestComposeModel(t, sendmail.NewSMTPMailer())
	makeComposeExternal(&model)
	model.selectedFile = "/path/that/does/not/exist.txt"
	model, _ = updateComposeModel(t, model, ctrlKey('s'))
	model.pendingConfirmation.input.SetValue("SEND")

	_, cmd := updateComposeModel(t, model, tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("confirmed send did not create a command")
	}
	result, ok := cmd().(sendMailProcess)
	if !ok {
		t.Fatalf("command result type = %T, want sendMailProcess", result)
	}
	if result.err == nil || !strings.Contains(result.err.Error(), model.selectedFile) {
		t.Fatalf("error = %v, want missing attachment path", result.err)
	}
	if smtpCalls != 0 {
		t.Fatalf("SMTP calls = %d, want 0", smtpCalls)
	}
}

func newSafetyTestComposeModel(t *testing.T, mailer mail.Mailer) ComposeModel {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)

	fields := []textinput.Model{
		newHeaderInput("FROM"),
		newHeaderInput("TO"),
		newHeaderInput("CC"),
		newHeaderInput("BCC"),
		newHeaderInput("SUBJECT"),
		newHeaderInput("HOST"),
		newHeaderInput("DEFAULT IS 25"),
	}
	values := []string{
		"weitingshih@rd01.softnext.com.tw",
		"jllee@rd01.softnext.com.tw",
		"",
		"",
		"Safety test",
		"smtp-test.example",
		"25",
	}
	for index, value := range values {
		fields[index].SetValue(value)
	}
	composer := textarea.New()
	composer.SetValue("body")
	preview := viewport.New(viewport.WithWidth(40), viewport.WithHeight(20))
	preview.SetContent("body")

	return ComposeModel{
		mailFields: fields,
		composer:   composer,
		preview:    preview,
		width:      100,
		height:     30,
		mailer:     mailer,
	}
}

func makeComposeExternal(model *ComposeModel) {
	model.mailFields[0].SetValue("sender@example.com")
	model.mailFields[1].SetValue("to@example.com")
	model.mailFields[2].SetValue("cc@example.net")
	model.mailFields[3].SetValue("bcc@example.org")
	model.mailFields[6].SetValue("1025")
}

func updateComposeModel(t *testing.T, model ComposeModel, msg tea.Msg) (ComposeModel, tea.Cmd) {
	t.Helper()
	updated, cmd := model.Update(msg)
	compose, ok := updated.(ComposeModel)
	if !ok {
		t.Fatalf("Update() model type = %T, want ComposeModel", updated)
	}
	return compose, cmd
}

func ctrlKey(char rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: char, Mod: tea.ModCtrl}
}

func textKey(char rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: char, Text: string(char)}
}
