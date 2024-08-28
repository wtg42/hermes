// Generate TUI component functions,
// mostly like dialogs, forms, and buttons.
// These are components we use a lot.
package tui

import (
	"go-go-power-mail/utils"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// form 的按鈕 被 getFormLayout 使用
func getFormButton() string {
	enterButtonStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#FF4D94")).
		Foreground(lipgloss.Color("#FFFFFF")). // 這個顏色好像沒有顯示出來
		Padding(0, 2).
		MarginRight(2)

	cancelButtonStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#878B7D")).
		Foreground(lipgloss.Color("#FFFFFF")). // 這個顏色好像沒有顯示出來
		Padding(0, 2)

	enterButton := enterButtonStyle.Render("確定[Enter]")
	cancelButton := cancelButtonStyle.Render("取消[Esc]")

	formButtonRow := lipgloss.JoinHorizontal(lipgloss.Left, enterButton, cancelButton)
	w, _ := utils.GetWindowSize()
	alignedRow := lipgloss.NewStyle().Width(w / 2).Align(lipgloss.Center).Render(formButtonRow)

	return alignedRow
}

// 產生 dialog layout 最後確認用
func getDialogBuilder(description string) strings.Builder {
	width, height := utils.GetWindowSize()
	doc := strings.Builder{}

	{
		var subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
		dialogBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

		buttonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			Padding(0, 3).
			MarginTop(1)

		// 之後可以支援 mouse event 可以加上底線效果
		activeButtonStyle := buttonStyle.
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#F25D94")).
			MarginRight(2)

		okButton := activeButtonStyle.Render("Yes[Enter]")
		cancelButton := buttonStyle.Render("No[Esc]")

		question := lipgloss.
			NewStyle().
			Width(50).
			Align(lipgloss.Center).
			Render(description)

		buttons := lipgloss.JoinHorizontal(lipgloss.Center, okButton, cancelButton)
		ui := lipgloss.JoinVertical(lipgloss.Center, question, buttons)

		dialog := lipgloss.Place(width, height,
			lipgloss.Center, lipgloss.Center,
			dialogBoxStyle.Render(ui),
			lipgloss.WithWhitespaceForeground(subtle),
		)

		doc.WriteString(dialog + "\n\n")
	}
	return doc
}

// 產生 alert layout
func getAlertBuilder(description ...string) strings.Builder {
	question := lipgloss.
		NewStyle().
		Width(50).
		Align(lipgloss.Center).
		Render(strings.Join(description, "\n"))

	dialogBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)

	ui := lipgloss.JoinVertical(lipgloss.Center, question)
	width, height := utils.GetWindowSize()
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
