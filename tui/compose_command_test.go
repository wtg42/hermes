package tui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestComposePrefixKeysDoNotReachInputs(t *testing.T) {
	tests := []struct {
		name        string
		activePanel int
		value       func(ComposeModel) string
	}{
		{
			name:        "header",
			activePanel: 0,
			value:       func(m ComposeModel) string { return m.mailFields[0].Value() },
		},
		{
			name:        "composer",
			activePanel: 1,
			value:       func(m ComposeModel) string { return m.composer.Value() },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newCommandTestComposeModel()
			m.activePanel = tt.activePanel
			if tt.activePanel == 1 {
				m.mailFields[0].Blur()
				m.composer.Focus()
			}

			m, _ = updateCompose(t, m, ctrlKey('x'))
			m, _ = updateCompose(t, m, runeKey('z'))

			if got := tt.value(m); got != "" {
				t.Fatalf("expected command key not to reach input, got %q", got)
			}
		})
	}
}

func TestComposeDirtyStateIncludesHeaderBodyAndAttachment(t *testing.T) {
	tests := []struct {
		name  string
		dirty func(*ComposeModel)
	}{
		{"header", func(m *ComposeModel) { m.mailFields[0].SetValue("sender@example.com") }},
		{"body", func(m *ComposeModel) { m.composer.SetValue("body") }},
		{"attachment", func(m *ComposeModel) { m.selectedFile = "/tmp/report.txt" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newCommandTestComposeModel()
			tt.dirty(&m)

			if !m.isDirty() {
				t.Fatal("expected compose to be dirty")
			}
			m, _ = updateCompose(t, m, ctrlKey('x'))
			m, _ = updateCompose(t, m, runeKey('c'))
			if m.prefix.mode != commandModeConfirmClear {
				t.Fatalf("expected clear confirmation, got %v", m.prefix.mode)
			}
		})
	}
}

func TestComposeClearConfirmation(t *testing.T) {
	m := newCommandTestComposeModel()
	m.mailFields[0].SetValue("sender@example.com")
	m.composer.SetValue("body")
	m.preview.SetContent("body")
	m.selectedFile = "/tmp/report.txt"

	m, _ = updateCompose(t, m, ctrlKey('x'))
	m, _ = updateCompose(t, m, runeKey('c'))
	m, _ = updateCompose(t, m, specialKey(tea.KeyEscape))
	if !m.isDirty() {
		t.Fatal("expected Esc to preserve dirty compose")
	}

	m, _ = updateCompose(t, m, ctrlKey('x'))
	m, _ = updateCompose(t, m, runeKey('c'))
	m, _ = updateCompose(t, m, runeKey('c'))

	if m.isDirty() {
		t.Fatal("expected confirmed clear to reset compose")
	}
	if strings.TrimSpace(m.preview.View()) != "" {
		t.Fatalf("expected preview to be cleared, got %q", m.preview.View())
	}
}

func TestComposeQuitFlow(t *testing.T) {
	t.Run("blank compose quits immediately", func(t *testing.T) {
		m := newCommandTestComposeModel()
		m, _ = updateCompose(t, m, ctrlKey('x'))
		_, cmd := updateCompose(t, m, runeKey('q'))
		assertQuitCommand(t, cmd)
	})

	t.Run("dirty compose requires repeated q", func(t *testing.T) {
		m := newCommandTestComposeModel()
		m.composer.SetValue("draft")
		m, _ = updateCompose(t, m, ctrlKey('x'))
		m, cmd := updateCompose(t, m, runeKey('q'))
		if cmd != nil {
			t.Fatal("expected first q not to quit dirty compose")
		}
		if m.prefix.mode != commandModeConfirmQuit {
			t.Fatalf("expected quit confirmation, got %v", m.prefix.mode)
		}

		_, cmd = updateCompose(t, m, runeKey('q'))
		assertQuitCommand(t, cmd)
	})

	t.Run("escape cancels dirty quit", func(t *testing.T) {
		m := newCommandTestComposeModel()
		m.composer.SetValue("draft")
		m, _ = updateCompose(t, m, ctrlKey('x'))
		m, _ = updateCompose(t, m, runeKey('q'))
		m, cmd := updateCompose(t, m, specialKey(tea.KeyEscape))

		if cmd != nil {
			t.Fatal("expected Esc not to quit")
		}
		if m.prefix.mode != commandModeNormal || m.composer.Value() != "draft" {
			t.Fatal("expected Esc to restore normal state and preserve draft")
		}
	})
}

func TestComposeEscAndCtrlCAreSafeInNormalState(t *testing.T) {
	for _, msg := range []tea.KeyPressMsg{specialKey(tea.KeyEscape), ctrlKey('c')} {
		m := newCommandTestComposeModel()
		m.composer.SetValue("draft")

		updated, cmd := updateCompose(t, m, msg)

		if cmd != nil {
			t.Fatalf("expected %s not to quit", msg.String())
		}
		if updated.composer.Value() != "draft" {
			t.Fatalf("expected %s to preserve body", msg.String())
		}
	}
}

func TestComposeCtrlCDoesNotQuitSafetyConfirmation(t *testing.T) {
	m := newCommandTestComposeModel()
	m.pendingConfirmation = &sendConfirmation{input: textinput.New()}

	updated, cmd := updateCompose(t, m, ctrlKey('c'))

	if cmd != nil {
		t.Fatal("expected Ctrl+C not to quit from safety confirmation")
	}
	if updated.pendingConfirmation == nil {
		t.Fatal("expected safety confirmation to remain open")
	}
}

func TestComposePrefixAttachAndFilepickerEscape(t *testing.T) {
	m := newCommandTestComposeModel()
	m.composer.SetValue("draft")
	m, _ = updateCompose(t, m, ctrlKey('x'))
	m, _ = updateCompose(t, m, runeKey('a'))

	if !m.showFilePicker {
		t.Fatal("expected prefix attach to open filepicker")
	}

	m, _ = updateCompose(t, m, specialKey(tea.KeyEscape))
	if m.showFilePicker {
		t.Fatal("expected Esc to close filepicker")
	}
	if m.composer.Value() != "draft" {
		t.Fatal("expected closing filepicker to preserve draft")
	}
}

func TestComposeTemplatePrefix(t *testing.T) {
	tests := []struct {
		key       rune
		wantValue string
	}{
		{'h', htmlTemplate},
		{'p', textTemplate},
		{'e', emlTemplate},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			m := newCommandTestComposeModel()
			m.activePanel = 1
			m.composer.Focus()

			m, _ = updateCompose(t, m, ctrlKey('x'))
			m, _ = updateCompose(t, m, runeKey('t'))
			m, _ = updateCompose(t, m, runeKey(tt.key))

			if m.composer.Value() != tt.wantValue {
				t.Fatalf("unexpected template body: %q", m.composer.Value())
			}
			if !strings.Contains(m.preview.View(), strings.Split(tt.wantValue, "\n")[0]) {
				t.Fatalf("expected preview to contain template, got %q", m.preview.View())
			}
		})
	}
}

func TestComposeTemplateEscapeReturnsToRoot(t *testing.T) {
	m := newCommandTestComposeModel()
	m, _ = updateCompose(t, m, ctrlKey('x'))
	m, _ = updateCompose(t, m, runeKey('t'))
	m, _ = updateCompose(t, m, specialKey(tea.KeyEscape))

	if m.prefix.mode != commandModeRoot {
		t.Fatalf("expected template Esc to return to root, got %v", m.prefix.mode)
	}
}

func TestComposeHelpOverlayDoesNotLeakKeys(t *testing.T) {
	m := newCommandTestComposeModel()
	m.activePanel = 1
	m.composer.Focus()
	m, _ = updateCompose(t, m, ctrlKey('x'))
	m, _ = updateCompose(t, m, runeKey('?'))

	if m.prefix.mode != commandModeHelp {
		t.Fatalf("expected help mode, got %v", m.prefix.mode)
	}

	m, _ = updateCompose(t, m, runeKey('x'))
	if m.composer.Value() != "" {
		t.Fatalf("expected help key not to reach composer, got %q", m.composer.Value())
	}

	m, _ = updateCompose(t, m, specialKey(tea.KeyEscape))
	if m.prefix.mode != commandModeNormal {
		t.Fatalf("expected Esc to close help, got %v", m.prefix.mode)
	}
}

func TestComposeHelpOverlayRendersCommands(t *testing.T) {
	m := newCommandTestComposeModel()
	m.prefix.mode = commandModeHelp

	content := ansi.Strip(m.View().Content)

	for _, want := range []string{"Commands", "Ctrl+S", "Ctrl+X", "[A] Attach", "[Q] Quit", "[Esc] Close"} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected help overlay to contain %q, got %q", want, content)
		}
	}
}

func TestComposeStatusBarModes(t *testing.T) {
	tests := []struct {
		name     string
		mode     commandMode
		notice   string
		contains []string
	}{
		{"normal", commandModeNormal, "", []string{"[Ctrl+S] Send", "[Ctrl+X] Commands", "[Tab] Next Field"}},
		{"root", commandModeRoot, "", []string{"COMMAND", "[A] Attach", "[Q] Quit", "[Esc] Cancel"}},
		{"template", commandModeTemplate, "", []string{"TEMPLATE", "[H] HTML", "[P] Plain Text", "[Esc] Back"}},
		{"confirm clear", commandModeConfirmClear, "", []string{"CLEAR", "[C] Confirm", "[Esc] Cancel"}},
		{"confirm quit", commandModeConfirmQuit, "", []string{"QUIT", "[Q] Confirm", "[Esc] Cancel"}},
		{"unknown", commandModeNormal, "Unknown command: z", []string{"Unknown command: z"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newCommandTestComposeModel()
			m.prefix.mode = tt.mode
			m.prefix.notice = tt.notice

			status := ansi.Strip(m.renderStatusBar())
			for _, want := range tt.contains {
				if !strings.Contains(status, want) {
					t.Fatalf("expected status %q to contain %q", status, want)
				}
			}
		})
	}
}

func TestComposeDirectNavigationAndSendingStatusRemain(t *testing.T) {
	m := newCommandTestComposeModel()
	m, _ = updateCompose(t, m, ctrlKey('j'))
	if m.activePanel != 1 {
		t.Fatal("expected Ctrl+J to focus composer")
	}
	m, _ = updateCompose(t, m, ctrlKey('k'))
	if m.activePanel != 0 {
		t.Fatal("expected Ctrl+K to focus header")
	}
	m, _ = updateCompose(t, m, specialKey(tea.KeyTab))
	if m.focusedField != 1 {
		t.Fatal("expected Tab to advance header field")
	}

	m.mailFields[5].SetValue("smtp.example.com")
	m.mailFields[6].SetValue("587")
	if status := ansi.Strip(m.renderStatusBar()); !strings.Contains(status, "SMTP target smtp.example.com:587") || strings.Contains(status, "TLS active") || strings.Contains(status, "Connected") {
		t.Fatalf("expected truthful SMTP target status, got %q", status)
	}

	// 使用 safe-send whitelist 資料驗證 Ctrl+S 仍是 direct shortcut；
	// 越界資料會正確地先進入安全確認，而不會立即開始寄送。
	m.mailFields[0].SetValue("weitingshih@rd01.softnext.com.tw")
	m.mailFields[1].SetValue("recipient@rd01.softnext.com.tw")
	m.mailFields[6].SetValue("25")
	m, sendCmd := updateCompose(t, m, ctrlKey('s'))
	if !m.sending || sendCmd == nil {
		t.Fatal("expected Ctrl+S to preserve direct send behavior")
	}

	m.sending = true
	if status := ansi.Strip(m.renderStatusBar()); !strings.Contains(status, "Sending") {
		t.Fatalf("expected sending status, got %q", status)
	}
}

func TestLegacyDirectKeysDoNotTriggerAppCommands(t *testing.T) {
	m := newCommandTestComposeModel()
	m.activePanel = 1
	m.composer.Focus()

	for _, msg := range []tea.KeyPressMsg{ctrlKey('a'), ctrlKey('h'), ctrlKey('t'), ctrlKey('e')} {
		var cmd tea.Cmd
		m, cmd = updateCompose(t, m, msg)
		if cmd != nil {
			_ = cmd
		}
		if m.showFilePicker {
			t.Fatalf("expected %s not to open filepicker", msg.String())
		}
		if m.composer.Value() == htmlTemplate || m.composer.Value() == textTemplate || m.composer.Value() == emlTemplate {
			t.Fatalf("expected %s not to apply template", msg.String())
		}
	}
}

func newCommandTestComposeModel() ComposeModel {
	fields := make([]textinput.Model, 7)
	for i := range fields {
		fields[i] = newHeaderInput([]string{"FROM", "TO", "CC", "BCC", "SUBJECT", "HOST", "DEFAULT IS 25"}[i])
	}
	fields[0].Focus()

	composer := textarea.New()
	preview := viewport.New(viewport.WithWidth(50), viewport.WithHeight(20))
	fp := filepicker.New()

	return ComposeModel{
		mailFields:   fields,
		focusedField: 0,
		composer:     composer,
		preview:      preview,
		filepicker:   fp,
		activePanel:  0,
		width:        120,
		height:       24,
		prefix:       newCommandPrefix(),
	}
}

func updateCompose(t *testing.T, m ComposeModel, msg tea.Msg) (ComposeModel, tea.Cmd) {
	t.Helper()
	updated, cmd := m.Update(msg)
	compose, ok := updated.(ComposeModel)
	if !ok {
		t.Fatalf("expected ComposeModel, got %T", updated)
	}
	return compose, cmd
}

func assertQuitCommand(t *testing.T, cmd tea.Cmd) {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", cmd())
	}
}
