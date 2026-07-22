package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestCommandPrefixStateTransitions(t *testing.T) {
	tests := []struct {
		name       string
		keys       []tea.KeyPressMsg
		wantMode   commandMode
		wantNotice string
	}{
		{
			name:     "ctrl+x enters root prefix",
			keys:     []tea.KeyPressMsg{ctrlKey('x')},
			wantMode: commandModeRoot,
		},
		{
			name:     "escape cancels root prefix",
			keys:     []tea.KeyPressMsg{ctrlKey('x'), specialKey(tea.KeyEscape)},
			wantMode: commandModeNormal,
		},
		{
			name:       "unknown key cancels root prefix",
			keys:       []tea.KeyPressMsg{ctrlKey('x'), runeKey('z')},
			wantMode:   commandModeNormal,
			wantNotice: "Unknown command",
		},
		{
			name:     "escape returns from template to root",
			keys:     []tea.KeyPressMsg{ctrlKey('x'), runeKey('t'), specialKey(tea.KeyEscape)},
			wantMode: commandModeRoot,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix := newCommandPrefix()
			for _, msg := range tt.keys {
				prefix, _, _ = prefix.Update(msg)
			}

			if prefix.mode != tt.wantMode {
				t.Fatalf("expected mode %v, got %v", tt.wantMode, prefix.mode)
			}
			if !strings.Contains(prefix.notice, tt.wantNotice) {
				t.Fatalf("expected notice %q, got %q", tt.wantNotice, prefix.notice)
			}
		})
	}
}

func TestCommandRegistryDrivesDispatchAndHUD(t *testing.T) {
	tests := []struct {
		mode      commandMode
		key       string
		wantID    commandID
		wantLabel string
	}{
		{commandModeRoot, "a", commandAttach, "Attach"},
		{commandModeRoot, "t", commandTemplates, "Template"},
		{commandModeRoot, "p", commandPalette, "Palette"},
		{commandModeRoot, "c", commandClear, "Clear"},
		{commandModeRoot, "q", commandQuit, "Quit"},
		{commandModeRoot, "?", commandHelp, "Help"},
		{commandModeTemplate, "h", commandTemplateHTML, "HTML"},
		{commandModeTemplate, "p", commandTemplateText, "Plain Text"},
		{commandModeTemplate, "e", commandTemplateEML, "EML"},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String()+"/"+tt.key, func(t *testing.T) {
			def, ok := lookupCommand(tt.mode, tt.key)
			if !ok {
				t.Fatalf("expected command for mode %v key %q", tt.mode, tt.key)
			}
			if def.ID != tt.wantID {
				t.Fatalf("expected command %q, got %q", tt.wantID, def.ID)
			}
			if def.Label != tt.wantLabel {
				t.Fatalf("expected label %q, got %q", tt.wantLabel, def.Label)
			}

			hud := renderCommandHUD(tt.mode)
			if !strings.Contains(hud, "["+strings.ToUpper(tt.key)+"] "+tt.wantLabel) {
				t.Fatalf("expected HUD to include registry command, got %q", hud)
			}
		})
	}
}

func TestCommandPrefixDoesNotTimeoutOnNonKeyMessage(t *testing.T) {
	prefix := newCommandPrefix()
	prefix, _, _ = prefix.Update(ctrlKey('x'))

	updated, command, handled := prefix.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	if updated.mode != commandModeRoot {
		t.Fatalf("expected root prefix to remain active, got %v", updated.mode)
	}
	if command != commandNone {
		t.Fatalf("expected no command for non-key message, got %q", command)
	}
	if handled {
		t.Fatal("expected non-key message to remain available to the active model")
	}
}

func runeKey(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Text: string(code)}
}

func specialKey(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}
