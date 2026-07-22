package tui

import (
	"image/color"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

func TestComposeApplyThemeStylesEveryComponent(t *testing.T) {
	m := newCommandTestComposeModel()
	theme, err := ResolveTheme("tokyo-night")
	if err != nil {
		t.Fatal(err)
	}

	m.applyTheme(theme)

	if m.theme.Name != "tokyo-night" {
		t.Fatalf("current theme = %q, want tokyo-night", m.theme.Name)
	}
	assertStyleColor(t, "canvas background", m.canvasStyle().GetBackground(), theme.Canvas)

	headerStyles := m.mailFields[0].Styles()
	assertStyleColor(t, "textinput focused text", headerStyles.Focused.Text.GetForeground(), theme.Text)
	assertStyleColor(t, "textinput focused background", headerStyles.Focused.Text.GetBackground(), theme.Panel)
	assertStyleColor(t, "textinput placeholder", headerStyles.Focused.Placeholder.GetForeground(), theme.Placeholder)
	assertStyleColor(t, "textinput cursor", headerStyles.Cursor.Color, theme.Accent)

	composerStyles := m.composer.Styles()
	assertStyleColor(t, "textarea base background", composerStyles.Focused.Base.GetBackground(), theme.Panel)
	assertStyleColor(t, "textarea text", composerStyles.Focused.Text.GetForeground(), theme.Text)
	assertStyleColor(t, "textarea placeholder", composerStyles.Focused.Placeholder.GetForeground(), theme.Placeholder)
	assertStyleColor(t, "textarea selection", composerStyles.Focused.CursorLine.GetBackground(), theme.Selection)
	assertStyleColor(t, "textarea cursor", composerStyles.Cursor.Color, theme.Accent)

	assertStyleColor(t, "preview background", m.preview.Style.GetBackground(), theme.Panel)
	assertStyleColor(t, "preview text", m.preview.Style.GetForeground(), theme.Text)
	assertStyleColor(t, "preview border", m.preview.Style.GetBorderTopForeground(), theme.Border)

	assertStyleColor(t, "filepicker file", m.filepicker.Styles.File.GetForeground(), theme.Text)
	assertStyleColor(t, "filepicker selected", m.filepicker.Styles.Selected.GetForeground(), theme.Accent)
}

func TestComposeApplyThemePreservesComposeState(t *testing.T) {
	m := newCommandTestComposeModel()
	m.mailFields[0].SetValue("sender@example.com")
	m.composer.SetValue("draft body")
	m.preview.SetContent("draft body")
	m.selectedFile = "/tmp/report.txt"
	m.activePanel = 1
	m.focusedField = 3

	theme, _ := ResolveTheme("tokyo-night")
	m.applyTheme(theme)

	if got := m.mailFields[0].Value(); got != "sender@example.com" {
		t.Fatalf("header changed after theme switch: %q", got)
	}
	if got := m.composer.Value(); got != "draft body" {
		t.Fatalf("body changed after theme switch: %q", got)
	}
	if got := m.selectedFile; got != "/tmp/report.txt" {
		t.Fatalf("attachment changed after theme switch: %q", got)
	}
	if m.activePanel != 1 || m.focusedField != 3 {
		t.Fatalf("focus changed after theme switch: panel=%d field=%d", m.activePanel, m.focusedField)
	}
}

func TestComposeViewFillsCanvasAndPanelsWithCurrentTheme(t *testing.T) {
	originalProfile := lipgloss.Writer.Profile
	lipgloss.Writer.Profile = colorprofile.TrueColor
	t.Cleanup(func() { lipgloss.Writer.Profile = originalProfile })

	tests := []struct {
		name      string
		canvasSGR string
		panelSGR  string
	}{
		{name: "gruvbox", canvasSGR: "48;2;40;40;40", panelSGR: "48;2;60;56;54"},
		{name: "tokyo-night", canvasSGR: "48;2;26;27;38", panelSGR: "48;2;36;40;59"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newCommandTestComposeModel()
			theme, _ := ResolveTheme(tt.name)
			m.applyTheme(theme)

			view := m.View()
			content := view.Content
			for _, want := range []string{tt.canvasSGR, tt.panelSGR} {
				if !strings.Contains(content, want) {
					t.Fatalf("%s View() does not contain color sequence %q", tt.name, want)
				}
			}
			assertStyleColor(t, "View background", view.BackgroundColor, theme.Canvas)
			assertStyleColor(t, "View foreground", view.ForegroundColor, theme.Text)
		})
	}
}

func TestComposeViewDropsPreviousThemeColorsAfterSwitch(t *testing.T) {
	originalProfile := lipgloss.Writer.Profile
	lipgloss.Writer.Profile = colorprofile.TrueColor
	t.Cleanup(func() { lipgloss.Writer.Profile = originalProfile })

	m := newCommandTestComposeModel()
	tokyoNight, _ := ResolveTheme("tokyo-night")
	m.applyTheme(tokyoNight)

	content := m.View().Content
	if !strings.Contains(content, "48;2;26;27;38") {
		t.Fatal("Tokyo Night canvas background is not rendered")
	}
	if strings.Contains(content, "48;2;40;40;40") {
		t.Fatal("Gruvbox canvas color remained after switching to Tokyo Night")
	}
}

func assertStyleColor(t *testing.T, label string, got color.Color, want string) {
	t.Helper()
	wantColor := lipgloss.Color(want)
	gotR, gotG, gotB, gotA := got.RGBA()
	wantR, wantG, wantB, wantA := wantColor.RGBA()
	if gotR != wantR || gotG != wantG || gotB != wantB || gotA != wantA {
		t.Fatalf("%s = rgba(%d,%d,%d,%d), want %s", label, gotR, gotG, gotB, gotA, want)
	}
}
