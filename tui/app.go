package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	errMsg error
)

type AppModel struct {
	SMTPIP  textinput.Model // 用戶輸入的 SMTP IP
	Focused int             // 當前焦點的位置
	err     error
}

func NewAppModel() AppModel {
	ti := textinput.New()
	ti.Placeholder = "輸入 SMTP IP"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return AppModel{
		SMTPIP:  ti,
		Focused: 0,
	}
}

func (m AppModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return textinput.Blink
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.SMTPIP, cmd = m.SMTPIP.Update(msg)
	return m, cmd
}

func (m AppModel) View() string {
	return m.SMTPIP.View() + "\n"
}

// func Start() {
// 	p := tea.NewProgram(initialModel())
// 	if _, err := p.Run(); err != nil {
// 		fmt.Printf("Alas, there's been an error: %v", err)
// 		os.Exit(1)
// 	}
// }
