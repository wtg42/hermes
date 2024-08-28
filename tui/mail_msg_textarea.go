// textarea for the message in the email
// this is the final step
package tui

import (
	"hermes/sendmail"
	"hermes/utils"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

type MailMsgModel struct {
	textarea      textarea.Model
	previousModel MailFieldsModel
	filepicker    filepicker.Model
	selectedFile  string
	err           error
	menuIndex     int
}

type clearErrorMsg struct{}

func (m MailMsgModel) Init() tea.Cmd {
	// 如果是 menu 選擇要夾帶檔案 必須轉到 file-picker 畫面
	m.menuIndex = viper.Get("menu-index").(int)
	if m.menuIndex == 2 {
		return m.filepicker.Init()
	}

	return nil
}

func (m MailMsgModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case clearErrorMsg:
		m.err = nil
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

		// 資料都儲存完畢 紀錄當前狀態 之後可以跳轉回去顯示
		viper.Set("mail-fields-model", m.previousModel)

		return initAlertModel(warning), tea.ClearScreen
	}

	cmds := make([]tea.Cmd, 2)

	// When the user selects a function in the menu for attachment.
	if m.menuIndex == 2 {
		var fpCmd tea.Cmd
		m.filepicker, fpCmd = m.filepicker.Update(msg)
		cmds = append(cmds, fpCmd)

		// Did the user select a file?
		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			// Get the path of the selected file.
			m.selectedFile = path
		}
	}

	var cmd tea.Cmd
	m.textarea.Focus()
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m MailMsgModel) View() string {
	w, h := utils.GetWindowSize()
	var renderString string

	// Draw a box around the text area
	drawAEmptyBox(func(s lipgloss.Style) {
		if m.menuIndex == 2 {
			renderString = lipgloss.JoinVertical(lipgloss.Center, renderString, m.filepicker.View())
		}

		var submit string
		if m.textarea.Focused() {
			submit = normalStyle.Render("送出郵件")
		} else {
			submit = normalStyle.Render("👉 送出郵件")
		}

		var ui string
		if m.menuIndex == 2 {
			ui = lipgloss.JoinVertical(lipgloss.Center, m.textarea.View(), m.filepicker.View(), submit)
		} else {
			ui = lipgloss.JoinVertical(lipgloss.Center, m.textarea.View(), submit)
		}

		renderString = s.Render(ui)
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

	mmm := MailMsgModel{
		textarea:      ta,
		previousModel: m,
	}

	mmm.Init()

	return mmm
}
