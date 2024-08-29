// 寄信程式介面 收集用戶輸入在傳給 SMTP client 發送
package tui

import (
	"go-go-power-mail/sendmail"
	"go-go-power-mail/utils"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

// 主畫面 Model
type AppModel struct {
	MailFields []textinput.Model // 用戶輸入的欄位
	Focused    int               // 當前焦點的位置
	comfirm    bool              // 用戶最後確認
}

type sendMailProcess struct {
	result bool
	err    error
}

// 樣式集合宣告
var (
	focusedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DC851C")).
		Align(lipgloss.Left)
)

func InitialAppModel() AppModel {

	// AppModel.MailFields 數量初始化
	m := AppModel{
		MailFields: make([]textinput.Model, 7),
		comfirm:    false,
	}

	//
	for i := range m.MailFields {
		t := textinput.New()
		t.Cursor.Blink = true

		switch i {
		case 0:
			t.Placeholder = "From"
			t.CharLimit = 256
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "To"
			t.CharLimit = 512
		case 2:
			t.Placeholder = "Cc"
			t.CharLimit = 512
		case 3:
			t.Placeholder = "Bcc"
			t.CharLimit = 512
		case 4:
			t.Placeholder = "Subject"
			t.CharLimit = 256
		case 5:
			t.Placeholder = "Contents"
			t.CharLimit = 1024
		case 6:
			t.Placeholder = "Host"
			t.CharLimit = 64
		}

		m.MailFields[i] = t
	}

	return m
}

// 這個並不會被自動呼叫，因為他不是初始化的 model 你需要自行呼叫
func (m AppModel) Init() tea.Cmd {
	return nil
}

type UserInputModelValue struct {
	From     string
	To       string
	Cc       string
	Bcc      string
	Subject  string
	Contents string
	Host     string
}

func (m AppModel) getUseModelValue() UserInputModelValue {
	return UserInputModelValue{
		From:     m.MailFields[0].Value(),
		To:       m.MailFields[1].Value(),
		Cc:       m.MailFields[2].Value(),
		Bcc:      m.MailFields[3].Value(),
		Subject:  m.MailFields[4].Value(),
		Contents: m.MailFields[5].Value(),
		Host:     m.MailFields[6].Value(),
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c": // 這個畫面不要用字元跳出 因為使用者要輸入
			return m, tea.Quit

		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			// Cycle indexes
			if (s == "tab" || s == "down") && m.Focused < len(m.MailFields)-1 {
				m.Focused++
			}

			if (s == "shift+tab" || s == "up") && m.Focused > 0 {
				m.Focused--
			}

			// 樣式的更新
			cmds := make([]tea.Cmd, len(m.MailFields))
			for i := 0; i <= len(m.MailFields)-1; i++ {
				if i == m.Focused {
					// Set focused state
					cmds[i] = m.MailFields[i].Focus()
					m.MailFields[i].PromptStyle = focusedStyle
					m.MailFields[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.MailFields[i].Blur()
				m.MailFields[i].PromptStyle = lipgloss.NewStyle()
				m.MailFields[i].TextStyle = lipgloss.NewStyle()
			}
			return m, tea.Batch(cmds...)

		case "enter", "esc":
			s := msg.String()

			// If the value of m.comfirm is true, means the user has completed the form
			// program should process the send mail command
			// otherwise, the user just press the sned mail button
			if s == "enter" && m.comfirm {
				m.setMailFieldsToViper()

				resultChan := make(chan sendMailProcess, 1)

				// Send mail without blocking the main thread
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
			} else if s == "enter" && !m.comfirm {
				m.comfirm = true
				return m, nil
			}
			if s == "esc" && m.comfirm {
				m.comfirm = false
				return m, nil
			} else if s == "esc" && !m.comfirm {
				// Reset all fields
				for i := range m.MailFields {
					m.MailFields[i].SetValue("")
				}
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		return m, nil
	case sendMailProcess:
		// 直接開新畫面顯示好了 免得判斷太複雜
		// 開 Alert 元件來顯示結果
		var warning string
		if msg.err != nil {
			warning = "😩 " + msg.err.Error()
		} else {
			warning = "🎉 信件傳送成功"
		}
		return initAlertModel(warning), tea.ClearScreen
	}

	// Here will update the contents of user input if KeyMsg is not interrupted
	cmds := make([]tea.Cmd, len(m.MailFields))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.MailFields {
		m.MailFields[i], cmds[i] = m.MailFields[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

// MailFiels 的值都會儲存在 viper 中後 之後再寄信再取出
func (m AppModel) setMailFieldsToViper() {
	userInput := m.getUseModelValue()
	viper.Set("mailField.From", userInput.From)
	viper.Set("mailField.To", userInput.To)
	viper.Set("mailField.Cc", userInput.Cc)
	viper.Set("mailField.Bcc", userInput.Bcc)
	viper.Set("mailField.Subject", userInput.Subject)
	viper.Set("mailField.Contents", userInput.Contents)
	viper.Set("mailField.Host", userInput.Host)
}

func (m AppModel) View() string {
	// 表單按過確認就直接跳 dialog
	if m.comfirm {
		dialog := getDialogBuilder("確定送出嗎?")
		return dialog.String()
	}

	// Normally render the form
	return m.getFormLayout()
}

// 產生表單的畫面 讓用戶輸入信件訊息
func (m AppModel) getFormLayout() string {
	var b strings.Builder

	w, h := utils.GetWindowSize()

	// labels
	labels := []string{
		"寄件者: \n",
		"收件者: \n",
		"副本: \n",
		"密件副本: \n",
		"主旨: \n",
		"內容: \n",
		"信件主機: \n",
	}

	for i := range m.MailFields {
		inputFiledWithLabel := lipgloss.JoinVertical(
			lipgloss.Left,
			labels[i],
			m.MailFields[i].View(),
		)

		// 每個 input 都換行排版
		b.WriteString(inputFiledWithLabel + "\n\n")
	}

	inputFieldString := b.String()
	contents := lipgloss.JoinVertical(lipgloss.Left, inputFieldString, getFormButton())
	b.Reset()
	b.WriteString(contents)

	// 排版換行
	b.WriteString("\n")

	// 組合 form 外框
	formBoxStyle := lipgloss.NewStyle().
		Width(w/2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 1).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)

	formBox := formBoxStyle.Render(b.String())

	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, formBox)
}
