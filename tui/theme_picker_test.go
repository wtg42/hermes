package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestComposePrefixOpensThemePicker(t *testing.T) {
	m := newCommandTestComposeModel()
	m, _ = updateCompose(t, m, ctrlKey('x'))
	m, _ = updateCompose(t, m, runeKey('p'))

	if m.themePicker == nil {
		t.Fatal("expected Ctrl+X p to open Theme Picker")
	}
	if m.themePicker.originalTheme != "gruvbox" || m.currentTheme().Name != "gruvbox" {
		t.Fatalf("unexpected picker state: %+v theme=%q", m.themePicker, m.currentTheme().Name)
	}
}

func TestThemePickerPreviewsAndConfirmsTheme(t *testing.T) {
	m := newCommandTestComposeModel()
	m.mailFields[0].SetValue("sender@example.com")
	m.composer.SetValue("draft")
	m.selectedFile = "/tmp/report.txt"
	m.activePanel = 1
	m, _ = updateCompose(t, m, ctrlKey('x'))
	m, _ = updateCompose(t, m, runeKey('p'))

	m, _ = updateCompose(t, m, specialKey(tea.KeyDown))
	if got := m.currentTheme().Name; got != "tokyo-night" {
		t.Fatalf("preview theme = %q, want tokyo-night", got)
	}
	if m.mailFields[0].Value() != "sender@example.com" || m.composer.Value() != "draft" || m.selectedFile != "/tmp/report.txt" || m.activePanel != 1 {
		t.Fatal("theme preview changed Compose state")
	}

	m, _ = updateCompose(t, m, specialKey(tea.KeyEnter))
	if m.themePicker != nil {
		t.Fatal("expected Enter to close Theme Picker")
	}
	if got := m.currentTheme().Name; got != "tokyo-night" {
		t.Fatalf("confirmed theme = %q, want tokyo-night", got)
	}
}

func TestThemePickerEscapeRestoresOriginalTheme(t *testing.T) {
	m := newCommandTestComposeModel()
	m, _ = updateCompose(t, m, ctrlKey('x'))
	m, _ = updateCompose(t, m, runeKey('p'))
	m, _ = updateCompose(t, m, runeKey('j'))

	if got := m.currentTheme().Name; got != "tokyo-night" {
		t.Fatalf("preview theme = %q, want tokyo-night", got)
	}
	m, _ = updateCompose(t, m, specialKey(tea.KeyEscape))
	if m.themePicker != nil {
		t.Fatal("expected Esc to close Theme Picker")
	}
	if got := m.currentTheme().Name; got != "gruvbox" {
		t.Fatalf("theme after Esc = %q, want gruvbox", got)
	}
}

func TestThemePickerNavigationWrapsWithArrowsAndJK(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyPressMsg
		want string
	}{
		{name: "down", key: specialKey(tea.KeyDown), want: "tokyo-night"},
		{name: "j", key: runeKey('j'), want: "tokyo-night"},
		{name: "up wraps", key: specialKey(tea.KeyUp), want: "tokyo-night"},
		{name: "k wraps", key: runeKey('k'), want: "tokyo-night"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newCommandTestComposeModel()
			m, _ = updateCompose(t, m, ctrlKey('x'))
			m, _ = updateCompose(t, m, runeKey('p'))
			m, _ = updateCompose(t, m, tt.key)
			if got := m.currentTheme().Name; got != tt.want {
				t.Fatalf("theme = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestThemePickerIsolatesBackgroundShortcutsAndInput(t *testing.T) {
	m := newCommandTestComposeModel()
	m.activePanel = 1
	m.composer.Focus()
	m, _ = updateCompose(t, m, ctrlKey('x'))
	m, _ = updateCompose(t, m, runeKey('p'))

	m, sendCmd := updateCompose(t, m, ctrlKey('s'))
	if sendCmd != nil || m.sending || m.pendingConfirmation != nil {
		t.Fatal("Ctrl+S escaped Theme Picker")
	}
	m, _ = updateCompose(t, m, runeKey('x'))
	if m.composer.Value() != "" {
		t.Fatalf("Theme Picker key leaked into composer: %q", m.composer.Value())
	}
	if m.themePicker == nil {
		t.Fatal("unhandled key unexpectedly closed Theme Picker")
	}
}

func TestThemePickerViewListsOnlyBuiltInThemes(t *testing.T) {
	m := newCommandTestComposeModel()
	m, _ = updateCompose(t, m, ctrlKey('x'))
	m, _ = updateCompose(t, m, runeKey('p'))

	content := ansi.Strip(m.View().Content)
	for _, want := range []string{"Theme Picker", "gruvbox", "tokyo-night", "Enter", "Esc"} {
		if !strings.Contains(content, want) {
			t.Fatalf("Theme Picker View missing %q", want)
		}
	}
	if strings.Contains(content, "catppuccin") {
		t.Fatal("Theme Picker listed unsupported theme")
	}
}
