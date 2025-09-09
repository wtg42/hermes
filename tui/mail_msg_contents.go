// textarea for the message in the email
// filepicker for the attachments
// this is the final step
package tui

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	"github.com/wtg42/hermes/sendmail"
	"github.com/wtg42/hermes/utils"
)

// MailMsgModel 郵件內容與附件畫面模型
//   - whichOneOnFocus: 控制焦點於輸入區、檔案選擇或送出按鈕
type MailMsgModel struct {
	textarea        textarea.Model
	previousModel   MailFieldsModel
	filepicker      filepicker.Model
	selectedFile    string
	err             error
	whichOneOnFocus int // 1: textarea 2: filepicker 3: send button
}

// Email content templates
const (
	htmlTemplate = `<html>
<head>
    <title>Email Template</title>
</head>
<body>
    <h1>Hello!</h1>
    <p>This is an HTML email template.</p>
    <p>Best regards,<br>Your Name</p>
</body>
</html>`

	textTemplate = `Hello,

This is a plain text email template.

Best regards,
Your Name`

	emlTemplate = `Return-Path: <sender@example.com>
Received: by smtp.example.com id 123456; Mon, 1 Jan 2024 12:00:00 +0000
Date: Mon, 1 Jan 2024 12:00:00 +0000
From: Sender Name <sender@example.com>
To: Recipient Name <recipient@example.com>
Subject: Test Email
Content-Type: text/plain; charset=UTF-8

Hello,

This is a sample EML email content.

Best regards,
Sender Name`
)

type clearErrorMsg struct{}

// Init 初始化 MailMsgModel
func (m MailMsgModel) Init() tea.Cmd {
	return m.filepicker.Init()
}

// Update 處理使用者輸入與檔案選擇
func (m MailMsgModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case clearErrorMsg:
		m.err = nil
	case tea.KeyMsg:
		return m.keyMsgSwitcher(msg)
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

	// 私有的 Msg Type 只能靠字串分析
	isFilePickerReadDirMsg := fmt.Sprintf("%T", msg)
	if isFilePickerReadDirMsg == "filepicker.readDirMsg" {
		var fpCmd tea.Cmd
		m.filepicker, fpCmd = m.filepicker.Update(msg)

		// Did the user select a file?
		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			// Get the path of the selected file.
			m.selectedFile = path
		}

		return m, fpCmd
	}

	cmds := make([]tea.Cmd, 2)

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View 渲染郵件內容與附件畫面
func (m MailMsgModel) View() string {
	w, h, err := utils.GetWindowSize()
	if err != nil {
		log.Fatalf("Error getting terminal size: %v", err)
	}
	var renderString string

	// Draw a box around the text area
	drawAEmptyBox(func(s lipgloss.Style) {
		var submit string
		if m.whichOneOnFocus != 3 {
			submit = normalStyle.Render("送出郵件")
		} else {
			submit = normalStyle.Render("👉 送出郵件")
		}

		pickfileDscription := "\nPick a file: "
		if m.selectedFile != "" {
			pickfileDscription = "\nSelected file: " + m.filepicker.Styles.Selected.Render(m.selectedFile)
		}

		if m.whichOneOnFocus == 2 {
			pickfileDscription = focusedStyle.Render(pickfileDscription)
		}

		// Build the filepicker's UI
		var ui string
		ui = lipgloss.JoinVertical(lipgloss.Left, pickfileDscription, m.filepicker.View())
		ui = lipgloss.JoinVertical(lipgloss.Left, m.textarea.View(), ui)
		ui = lipgloss.JoinVertical(lipgloss.Center, ui, submit)

		renderString = s.Render(ui)
	})

	// Help text at the bottom of the message-box.
	renderString = lipgloss.JoinVertical(lipgloss.Left, renderString, "\nTab: Switch Focus • ctrl+c: Quit • Esc: Previous Page\nctrl+h: HTML Template • ctrl+t: Text Template • ctrl+e: EML Template\n")

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

// 初始化 MailMsgModel
// 這是給 mail field 去轉換畫面用的
// 回傳 Model 外，一併回傳 filepicker.Init() cmd 去執行才會有正確效果
func initMailMsgModel(m MailFieldsModel) (MailMsgModel, tea.Cmd) {
	// initialize textarea input
	ta := textarea.New()
	ta.Placeholder = "Add your message of mail here."
	ta.CharLimit = 0
	ta.SetWidth(50)
	ta.SetHeight(3)
	ta.Focus()

	mmm := MailMsgModel{
		textarea:        ta,
		previousModel:   m,
		whichOneOnFocus: 1,
	}

	fp := filepicker.New()
	fp.Height = 5
	// fp.AllowedTypes = []string{".mod", ".sum", ".go", ".txt", ".log"}
	var err error
	fp.CurrentDirectory, err = os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	mmm.filepicker = fp
	cmd := mmm.filepicker.Init() // 這個很重要 需要把指令回傳到主程式執行

	// mmm.Init()

	return mmm, cmd
}

func (m MailMsgModel) keyMsgSwitcher(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Behavior of textarea
	if m.whichOneOnFocus == 1 {
		switch msg.String() {
		case "esc":
			return InitialMailFieldsModel(), tea.ClearScreen
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+h":
			m.textarea.SetValue(htmlTemplate)
			return m, nil
		case "ctrl+t":
			m.textarea.SetValue(textTemplate)
			return m, nil
		case "ctrl+e":
			m.textarea.SetValue(emlTemplate)
			return m, nil
		case "tab":
			m.textarea.Blur()
			m.whichOneOnFocus = 2
			return m, nil
		default:
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}
	}

	// Behavior of filepicker
	if m.whichOneOnFocus == 2 {
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab":
			m.whichOneOnFocus = 3
			return m, nil
		default:
			var cmd tea.Cmd
			m.filepicker, cmd = m.filepicker.Update(msg)
			// Did the user select a file?
			if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
				// Get the path of the selected file.
				m.selectedFile = path
			}
			return m, cmd
		}
	}

	// Behavior of send button
	if m.whichOneOnFocus == 3 {
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab":
			m.whichOneOnFocus = 1
			cmd := m.textarea.Focus()
			return m, cmd
		case "enter":
			if !m.textarea.Focused() {
				// 記錄下來
				viper.Set("mailField.contents", m.textarea.Value())
				viper.Set("mailField.attachment", m.selectedFile)

				// Send the mail.
				m.previousModel.setMailFieldsToViper()
				return m.sendMailWithChannel()
			}
		}
	}
	return m, nil
}
