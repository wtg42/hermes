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

type AppModel struct {
	MailFields []textinput.Model // 用戶輸入的 SMTP IP
	Focused    int               // 當前焦點的位置
	err        error
}

func InitialAppModel() AppModel {

	// AppModel.MailFields 數量初始化
	m := AppModel{
		MailFields: make([]textinput.Model, 4),
	}

	//
	for i := range m.MailFields {
		t := textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Nickname"
			t.Focus()
			t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
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

func (m AppModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return textinput.Blink
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab", "shift+tab":
			s := msg.String()

			// Cycle indexes
			if s == "tab" {
				m.Focused++
			} else {
				m.Focused--
			}

			cmds := make([]tea.Cmd, len(m.MailFields))
			for i := 0; i <= len(m.MailFields)-1; i++ {
				if i == m.Focused {
					// Set focused state
					cmds[i] = m.MailFields[i].Focus()
					m.MailFields[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
					m.MailFields[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
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

	for i := range m.MailFields {
		b.WriteString(m.MailFields[i].View())
		if i < len(m.MailFields)-1 {
			b.WriteRune('\n')
		}
	}
	return b.String()
}

// func Start() {
// 	p := tea.NewProgram(initialModel())
// 	if _, err := p.Run(); err != nil {
// 		fmt.Printf("Alas, there's been an error: %v", err)
// 		os.Exit(1)
// 	}
// }
