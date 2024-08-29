// textarea for the message in the email
// this is the final step
package tui

import (
	"hermes/sendmail"
	"hermes/utils"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

type MailMsgModel struct {
	textarea      textarea.Model
	previousModel MailFieldsModel
}

func (m MailMsgModel) Init() tea.Cmd {
	return nil
}

func (m MailMsgModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab":
			m.textarea.Blur()
			return m, nil
		case "enter":
			if !m.textarea.Focused() {
				// Send the mail.
				viper.Set("mailField.contents", m.textarea.Value())
				m.previousModel.setMailFieldsToViper()
				return m.sendMailWithChannel()
			}
		}
	case sendMailProcess:
		// 直接開新畫面顯示好了 免得判斷太複雜
		// 開 Alert 元件來顯示結果
		var warning string
		if msg.err != nil {
			warning = "😩 " + msg.err.Error()
		} else {
			warning = "🎉 信件傳送成功"
		}
		viper.Set("mail-fields-model", m.previousModel)
		return initAlertModel(warning), tea.ClearScreen
	}
	var cmd tea.Cmd
	m.textarea.Focus()
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m MailMsgModel) View() string {
	w, h := utils.GetWindowSize()
	var renderString string

	drawAEmptyBox(func(s lipgloss.Style) {
		if m.textarea.Focused() {
			submit := normalStyle.Render("送出郵件")
			ui := lipgloss.JoinVertical(lipgloss.Center, m.textarea.View(), submit)
			renderString = s.Render(ui)
		} else {
			submit := normalStyle.Render("👉 送出郵件")
			ui := lipgloss.JoinVertical(lipgloss.Center, m.textarea.View(), submit)
			renderString = s.Render(ui)
		}
	})

	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, renderString)
}

// Asynchronously sends an email
func (m MailMsgModel) sendMailWithChannel() (tea.Model, tea.Cmd) {
	resultChan := make(chan sendMailProcess, 1)

	// Send mail without blocking the main thread
	// Bubbletea will trigger Update() by this message
	go func() {
		result, err := sendmail.SendMailWithMultipart("mailField")
		resultChan <- sendMailProcess{result: result, err: err}
	}()

	// We don't want to block the main thread,
	// so we wrap the channel with a func.
	// This Should return a tea.Msg to notify the main thread
	// that the send mail process is completed
	return m, func() tea.Msg {
		result := <-resultChan
		close(resultChan)
		return tea.Msg(result)
	}
}

func initMailMsgModel(m MailFieldsModel) MailMsgModel {
	// initialize textarea input
	ta := textarea.New()
	ta.Placeholder = "Add your message here."
	ta.CharLimit = 0
	ta.SetWidth(50)
	ta.SetHeight(3)
	ta.Focus()

	return MailMsgModel{
		textarea:      ta,
		previousModel: m,
	}
}
