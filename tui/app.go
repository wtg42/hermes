package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type AppModel struct {
	SMTPIP    textinput.Model // 用戶輸入的 SMTP IP
	Sender    textinput.Model // 用戶輸入的寄件者
	Recipient textinput.Model // 用戶輸入的收件者
	CC        textinput.Model // 用戶輸入的 CC
	BCC       textinput.Model // 用戶輸入的 BCC
	Subject   textinput.Model // 用戶輸入的主旨
	Body      textinput.Model // 用戶輸入的郵件內容
	Focused   int             // 當前焦點的位置
}

func NewAppModel() AppModel {
	return AppModel{
		SMTPIP:    textinput.New(),
		Sender:    textinput.New(),
		Recipient: textinput.New(),
		CC:        textinput.New(),
		BCC:       textinput.New(),
		Subject:   textinput.New(),
		Body:      textinput.New(),
		Focused:   0,
	}
}

func (m AppModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			fmt.Println("再見！")
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m AppModel) View() string {
	s := "What should we buy at the market?\n\n"
	return s
}

// func Start() {
// 	p := tea.NewProgram(initialModel())
// 	if _, err := p.Run(); err != nil {
// 		fmt.Printf("Alas, there's been an error: %v", err)
// 		os.Exit(1)
// 	}
// }
