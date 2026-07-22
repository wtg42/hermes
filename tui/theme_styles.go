package tui

import (
	"image/color"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

func (m *ComposeModel) applyTheme(theme Theme) {
	m.theme = theme

	for i := range m.mailFields {
		styles := m.mailFields[i].Styles()
		styles.Focused.Text = themedStyle(theme.Text, theme.Panel)
		styles.Focused.Placeholder = themedStyle(theme.Placeholder, theme.Panel)
		styles.Focused.Suggestion = themedStyle(theme.Info, theme.Panel)
		styles.Focused.Prompt = themedStyle(theme.Muted, theme.Panel)
		styles.Blurred.Text = themedStyle(theme.Muted, theme.Panel)
		styles.Blurred.Placeholder = themedStyle(theme.Placeholder, theme.Panel)
		styles.Blurred.Suggestion = themedStyle(theme.Muted, theme.Panel)
		styles.Blurred.Prompt = themedStyle(theme.Muted, theme.Panel)
		styles.Cursor.Color = lipgloss.Color(theme.Accent)
		m.mailFields[i].SetStyles(styles)
	}

	m.composer.SetStyles(textareaThemeStyles(m.composer.Styles(), theme))
	m.preview.Style = borderedPanelStyle(theme, theme.Border).
		Padding(0, 1)

	m.filepicker.Styles.DisabledCursor = themedStyle(theme.Muted, theme.Panel)
	m.filepicker.Styles.Cursor = themedStyle(theme.Accent, theme.Panel)
	m.filepicker.Styles.Symlink = themedStyle(theme.Info, theme.Panel)
	m.filepicker.Styles.Directory = themedStyle(theme.Info, theme.Panel)
	m.filepicker.Styles.File = themedStyle(theme.Text, theme.Panel)
	m.filepicker.Styles.DisabledFile = themedStyle(theme.Muted, theme.Panel)
	m.filepicker.Styles.Permission = themedStyle(theme.Muted, theme.Panel)
	m.filepicker.Styles.Selected = themedStyle(theme.Accent, theme.Selection).Bold(true)
	m.filepicker.Styles.DisabledSelected = themedStyle(theme.Muted, theme.Selection)
	m.filepicker.Styles.FileSize = themedStyle(theme.Muted, theme.Panel).Width(7).Align(lipgloss.Right)
	m.filepicker.Styles.EmptyDirectory = themedStyle(theme.Muted, theme.Panel).PaddingLeft(2).SetString("Bummer. No Files Found.")

	if m.pendingConfirmation != nil {
		styles := m.pendingConfirmation.input.Styles()
		styles = textInputThemeStyles(styles, theme)
		m.pendingConfirmation.input.SetStyles(styles)
	}
}

func (m ComposeModel) currentTheme() Theme {
	if m.theme.Name != "" {
		return m.theme
	}
	theme, _ := ResolveTheme(DefaultThemeName)
	return theme
}

func (m ComposeModel) canvasStyle() lipgloss.Style {
	theme := m.currentTheme()
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Text)).
		Background(lipgloss.Color(theme.Canvas))
}

func (m ComposeModel) panelStyle(focused bool) lipgloss.Style {
	theme := m.currentTheme()
	style := borderedPanelStyle(theme, theme.Border)
	if focused {
		style = style.BorderLeftForeground(lipgloss.Color(theme.Accent))
	}
	return style
}

func borderedPanelStyle(theme Theme, border string) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Text)).
		Background(lipgloss.Color(theme.Panel)).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(border)).
		BorderBackground(lipgloss.Color(theme.Canvas))
}

func textInputThemeStyles(styles textinput.Styles, theme Theme) textinput.Styles {
	styles.Focused.Text = themedStyle(theme.Text, theme.Panel)
	styles.Focused.Placeholder = themedStyle(theme.Placeholder, theme.Panel)
	styles.Focused.Suggestion = themedStyle(theme.Info, theme.Panel)
	styles.Focused.Prompt = themedStyle(theme.Muted, theme.Panel)
	styles.Blurred.Text = themedStyle(theme.Muted, theme.Panel)
	styles.Blurred.Placeholder = themedStyle(theme.Placeholder, theme.Panel)
	styles.Blurred.Suggestion = themedStyle(theme.Muted, theme.Panel)
	styles.Blurred.Prompt = themedStyle(theme.Muted, theme.Panel)
	styles.Cursor.Color = lipgloss.Color(theme.Accent)
	return styles
}

func textareaThemeStyles(styles textarea.Styles, theme Theme) textarea.Styles {
	styles.Focused.Base = themedStyle(theme.Text, theme.Panel)
	styles.Focused.Text = themedStyle(theme.Text, theme.Panel)
	styles.Focused.LineNumber = themedStyle(theme.Muted, theme.Panel)
	styles.Focused.CursorLineNumber = themedStyle(theme.Accent, theme.Selection)
	styles.Focused.CursorLine = themedStyle(theme.Text, theme.Selection)
	styles.Focused.EndOfBuffer = themedStyle(theme.Panel, theme.Panel)
	styles.Focused.Placeholder = themedStyle(theme.Placeholder, theme.Panel)
	styles.Focused.Prompt = themedStyle(theme.Muted, theme.Panel)
	styles.Blurred.Base = themedStyle(theme.Muted, theme.Panel)
	styles.Blurred.Text = themedStyle(theme.Muted, theme.Panel)
	styles.Blurred.LineNumber = themedStyle(theme.Muted, theme.Panel)
	styles.Blurred.CursorLineNumber = themedStyle(theme.Muted, theme.Panel)
	styles.Blurred.CursorLine = themedStyle(theme.Muted, theme.Panel)
	styles.Blurred.EndOfBuffer = themedStyle(theme.Panel, theme.Panel)
	styles.Blurred.Placeholder = themedStyle(theme.Placeholder, theme.Panel)
	styles.Blurred.Prompt = themedStyle(theme.Muted, theme.Panel)
	styles.Cursor.Color = lipgloss.Color(theme.Accent)
	return styles
}

func themedStyle(foreground, background string) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(foreground)).
		Background(lipgloss.Color(background))
}

// ensureOpaqueBackground fills only cells that do not already have an ANSI
// background. Explicit textarea styles such as the cursor-line selection are
// preserved while viewport padding is painted with the panel background.
func ensureOpaqueBackground(content string, width, height int, background color.Color) string {
	if width < 1 || height < 1 {
		return ""
	}

	buffer := uv.NewScreenBuffer(width, height)
	uv.NewStyledString(content).Draw(buffer, uv.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		coveredUntil := 0
		for x := 0; x < width; x++ {
			cell := buffer.CellAt(x, y)
			if x < coveredUntil {
				continue
			}
			if cell == nil || cell.IsZero() {
				cell = uv.EmptyCell.Clone()
				buffer.SetCell(x, y, cell)
			}
			if cell.Width > 1 {
				coveredUntil = x + cell.Width
			}
			if cell.Style.Bg == nil {
				cell.Style.Bg = background
			}
		}
	}
	return buffer.Render()
}
