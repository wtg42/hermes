package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	errMsg error
)

// 主畫面 Model
type AppModel struct {
	MailFields []textinput.Model // 用戶輸入的 SMTP IP
	Focused    int               // 當前焦點的位置
	err        error
}

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#DC851C"))
	// blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	// cursorStyle         = focusedStyle
	// noStyle             = lipgloss.NewStyle()
	// helpStyle           = blurredStyle
	// cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	// focusedButton = focusedStyle.Render("[ Submit ]")
	// blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

func InitialAppModel() AppModel {

	// AppModel.MailFields 數量初始化
	m := AppModel{
		MailFields: make([]textinput.Model, 4),
	}

	//
	for i := range m.MailFields {
		t := textinput.New()
		t.Cursor.Style = focusedStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Nickname"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Email"
			t.CharLimit = 64
		case 2:
			t.Placeholder = "To"
			t.CharLimit = 64
		case 3:
			t.Placeholder = "To"
			t.CharLimit = 64
		}

		m.MailFields[i] = t
	}

	ti := textinput.New()
	ti.Placeholder = "輸入 SMTP IP"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return m
}

// 這個並不會被自動呼叫，因為他不是初始化的 model 你需要自行呼叫
func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
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
		}
	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	// 訊息內如更新
	cmds := make([]tea.Cmd, len(m.MailFields))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.MailFields {
		m.MailFields[i], cmds[i] = m.MailFields[i].Update(msg)
	}
	return m, cmd
}

func (m AppModel) View() string {
	var b strings.Builder

	// labels
	labels := []string{"寄件者: \n", "收件者: \n", "主旨: \n", "內容: \n"}

	for i := range m.MailFields {
		b.WriteString(labels[i])
		b.WriteString(m.MailFields[i].View())
		if i < len(m.MailFields)-1 {
			b.WriteRune('\n')
		}
	}
	return b.String()
}
