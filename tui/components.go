// Generate TUI component functions,
// mostly like dialogs, forms, and buttons.
// These are components we use a lot.
package tui

import (
	"log"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/wtg42/hermes/utils"
)

// 樣式集合宣告（舊 EML flow 使用預設 Gruvbox theme）。
var focusedStyle = func() lipgloss.Style {
	theme, _ := ResolveTheme(DefaultThemeName)
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Accent)).
		Background(lipgloss.Color(theme.Canvas)).
		Align(lipgloss.Left)
}()

// 產生 alert layout
func getAlertBuilder(theme Theme, description ...string) strings.Builder {
	question := lipgloss.
		NewStyle().
		Width(50).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color(theme.Text)).
		Background(lipgloss.Color(theme.Panel)).
		Render(strings.Join(description, "\n"))

	dialogBoxStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Text)).
		Background(lipgloss.Color(theme.Panel)).
		BorderForeground(lipgloss.Color(theme.Accent)).
		Border(lipgloss.RoundedBorder()).
		Padding(1, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)

	ui := lipgloss.JoinVertical(lipgloss.Center, question)

	width, height, err := utils.GetWindowSize()
	if err != nil {
		log.Fatalf("Error getting terminal size: %v", err)
	}

	subtle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Muted)).
		Background(lipgloss.Color(theme.Canvas))
	alert := lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		dialogBoxStyle.Render(ui),
		lipgloss.WithWhitespaceStyle(subtle),
	)

	doc := strings.Builder{}

	doc.WriteString(alert)

	return doc
}

func drawAEmptyBox(callback func(s lipgloss.Style)) {
	theme, _ := ResolveTheme(DefaultThemeName)
	dialogBoxStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Text)).
		Background(lipgloss.Color(theme.Panel)).
		BorderForeground(lipgloss.Color(theme.Border)).
		Border(lipgloss.RoundedBorder()).
		Padding(1, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)

	callback(dialogBoxStyle)
}
