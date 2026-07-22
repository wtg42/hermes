package tui

import (
	"image/color"
	"reflect"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	uv "github.com/charmbracelet/ultraviolet"
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

func TestPanelStylesUseStraightBordersAndLeftFocusRail(t *testing.T) {
	for _, name := range AvailableThemes() {
		t.Run(name, func(t *testing.T) {
			m := newCommandTestComposeModel()
			theme, _ := ResolveTheme(name)
			m.applyTheme(theme)

			focused := m.panelStyle(true)
			if !reflect.DeepEqual(focused.GetBorderStyle(), lipgloss.NormalBorder()) {
				t.Fatalf("focused panel border = %#v, want NormalBorder", focused.GetBorderStyle())
			}
			assertStyleColor(t, "focused top border", focused.GetBorderTopForeground(), theme.Border)
			assertStyleColor(t, "focused right border", focused.GetBorderRightForeground(), theme.Border)
			assertStyleColor(t, "focused bottom border", focused.GetBorderBottomForeground(), theme.Border)
			assertStyleColor(t, "focused left rail", focused.GetBorderLeftForeground(), theme.Accent)
			assertBorderBackgrounds(t, "focused panel", focused, theme.Canvas)

			inactive := m.panelStyle(false)
			assertStyleColor(t, "inactive top border", inactive.GetBorderTopForeground(), theme.Border)
			assertStyleColor(t, "inactive right border", inactive.GetBorderRightForeground(), theme.Border)
			assertStyleColor(t, "inactive bottom border", inactive.GetBorderBottomForeground(), theme.Border)
			assertStyleColor(t, "inactive left border", inactive.GetBorderLeftForeground(), theme.Border)
			assertBorderBackgrounds(t, "inactive panel", inactive, theme.Canvas)

			if !reflect.DeepEqual(m.preview.Style.GetBorderStyle(), lipgloss.NormalBorder()) {
				t.Fatalf("preview border = %#v, want NormalBorder", m.preview.Style.GetBorderStyle())
			}
			assertStyleColor(t, "preview left border", m.preview.Style.GetBorderLeftForeground(), theme.Border)
			assertBorderBackgrounds(t, "preview", m.preview.Style, theme.Canvas)
		})
	}
}

func TestBorderedPanelUsesCanvasBorderAndPanelInterior(t *testing.T) {
	for _, name := range AvailableThemes() {
		t.Run(name, func(t *testing.T) {
			theme, _ := ResolveTheme(name)
			style := borderedPanelStyle(theme, theme.Border).
				Width(12).
				Height(5)

			assertStyleColor(t, "panel interior style", style.GetBackground(), theme.Panel)
			assertBorderBackgrounds(t, "bordered panel", style, theme.Canvas)

			rendered := style.Render("content")
			width, height := lipgloss.Width(rendered), lipgloss.Height(rendered)
			buffer := uv.NewScreenBuffer(width, height)
			uv.NewStyledString(rendered).Draw(buffer, uv.Rect(0, 0, width, height))

			for _, point := range [][2]int{
				{0, 0}, {width - 1, 0},
				{0, height - 1}, {width - 1, height - 1},
				{width / 2, 0}, {width / 2, height - 1},
				{0, height / 2}, {width - 1, height / 2},
			} {
				cell := buffer.CellAt(point[0], point[1])
				if cell == nil || !sameColor(cell.Style.Bg, lipgloss.Color(theme.Canvas)) {
					t.Fatalf("border cell (%d,%d) background = %v, want Canvas %s", point[0], point[1], cellBackground(cell), theme.Canvas)
				}
			}

			interior := buffer.CellAt(1, 1)
			if interior == nil || !sameColor(interior.Style.Bg, lipgloss.Color(theme.Panel)) {
				t.Fatalf("interior background = %v, want Panel %s", cellBackground(interior), theme.Panel)
			}
		})
	}
}

func TestComposerContentHasOpaqueThemeBackground(t *testing.T) {
	tests := []struct {
		name    string
		content string
		focused bool
	}{
		{name: "empty focused", focused: true},
		{name: "empty blurred"},
		{name: "partial focused", content: "draft body", focused: true},
		{name: "partial blurred", content: "draft body"},
	}

	for _, themeName := range AvailableThemes() {
		for _, tt := range tests {
			t.Run(themeName+"/"+tt.name, func(t *testing.T) {
				m := newCommandTestComposeModel()
				theme, _ := ResolveTheme(themeName)
				m.applyTheme(theme)
				m.composer.SetValue(tt.content)
				if tt.focused {
					m.composer.Focus()
				} else {
					m.composer.Blur()
				}

				content := m.renderComposerContent(36, 7)
				buffer := uv.NewScreenBuffer(36, 7)
				uv.NewStyledString(content).Draw(buffer, uv.Rect(0, 0, 36, 7))

				selectionCells := 0
				for y := 0; y < 7; y++ {
					for x := 0; x < 36; x++ {
						cell := buffer.CellAt(x, y)
						if cell == nil || cell.IsZero() {
							continue
						}
						if sameColor(cell.Style.Bg, lipgloss.Color(theme.Selection)) {
							selectionCells++
							continue
						}
						if !sameColor(cell.Style.Bg, lipgloss.Color(theme.Panel)) {
							t.Fatalf("cell (%d,%d) background = %v, want Panel %s or Selection %s", x, y, cell.Style.Bg, theme.Panel, theme.Selection)
						}
					}
				}
				if tt.focused && selectionCells == 0 {
					t.Fatal("focused textarea lost its Selection cursor-line background")
				}
			})
		}
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

func sameColor(got, want color.Color) bool {
	if got == nil || want == nil {
		return got == nil && want == nil
	}
	gotR, gotG, gotB, gotA := got.RGBA()
	wantR, wantG, wantB, wantA := want.RGBA()
	return gotR == wantR && gotG == wantG && gotB == wantB && gotA == wantA
}

func assertBorderBackgrounds(t *testing.T, label string, style lipgloss.Style, want string) {
	t.Helper()
	assertStyleColor(t, label+" top background", style.GetBorderTopBackground(), want)
	assertStyleColor(t, label+" right background", style.GetBorderRightBackground(), want)
	assertStyleColor(t, label+" bottom background", style.GetBorderBottomBackground(), want)
	assertStyleColor(t, label+" left background", style.GetBorderLeftBackground(), want)
}

func cellBackground(cell *uv.Cell) color.Color {
	if cell == nil {
		return nil
	}
	return cell.Style.Bg
}
