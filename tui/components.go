// Generate TUI component functions,
// mostly like dialogs, forms, and buttons.
// These are components we use a lot.
package tui

import (
	"log"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/wtg42/hermes/utils"
)

// 樣式集合宣告
var (
	focusedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DC851C")).
		Align(lipgloss.Left)
)

// 產生 alert layout
func getAlertBuilder(description ...string) strings.Builder {
	question := lipgloss.
		NewStyle().
		Width(50).
		Align(lipgloss.Center).
		Render(strings.Join(description, "\n"))

	dialogBoxStyle := lipgloss.NewStyle().
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

	var subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	alert := lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		dialogBoxStyle.Render(ui),
		lipgloss.WithWhitespaceForeground(subtle),
	)

	doc := strings.Builder{}

	doc.WriteString(alert)

	return doc
}

func drawAEmptyBox(callback func(s lipgloss.Style)) {
	dialogBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)

	callback(dialogBoxStyle)
}
