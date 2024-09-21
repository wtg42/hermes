// Generate TUI component functions,
// mostly like dialogs, forms, and buttons.
// These are components we use a lot.
package tui

import (
	"hermes/utils"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// 表單按鈕文字描述 顯示用戶目前游標
type FormButtonBuilder struct {
	Submit strings.Builder
	Cancel strings.Builder
}

// form 的按鈕 被 getFormLayout 使用
func (fb FormButtonBuilder) getFormButton(m MailFieldsModel) string {
	// style
	enterButtonStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#FF4D94")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 2).
		MarginRight(2)

	cancelButtonStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#878B7D")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 2)

	// Description
	buttonBuilder := FormButtonBuilder{
		Submit: strings.Builder{},
		Cancel: strings.Builder{},
	}

	buttonBuilder.Submit.WriteString("  下一步[Enter]")
	buttonBuilder.Cancel.WriteString("  取消[Esc]")

	// 根據 model 狀態改變按鈕
	switch {
	case m.ActiveFormSubmit:
		buttonBuilder.Submit.Reset()
		buttonBuilder.Submit.WriteString("👉下一步[Enter]")
	case m.ActiveFormCancel:
		buttonBuilder.Cancel.Reset()
		buttonBuilder.Cancel.WriteString("👉取消[Esc]")
	}

	enterButton := enterButtonStyle.Render(buttonBuilder.Submit.String())
	cancelButton := cancelButtonStyle.Render(buttonBuilder.Cancel.String())

	formButtonRow := lipgloss.JoinHorizontal(lipgloss.Left, enterButton, cancelButton)
	return formButtonRow
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

func drawAEmptyBox(callback func(s lipgloss.Style)) {
	dialogBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)

	callback(dialogBoxStyle)
}
