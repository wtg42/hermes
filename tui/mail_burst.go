package tui

import (
	"hermes/sendmail"
	"hermes/utils"
	"log"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MailBurstModel struct {
	numberTextInput textinput.Model
}

func (m MailBurstModel) Init() tea.Cmd {
	return nil
}

func (m MailBurstModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			quantity, err := strconv.ParseInt(m.numberTextInput.Value(), 10, 64)
			if err != nil {
				log.Fatalf("Failed to convert quantity to int: %+v", err)
			}
			sendmail.BurstModeSendMail(int(quantity), "localhost", "1025")
			return m, nil
		}
	}

	// 限制用戶只能輸入數字
	m.numberTextInput.SetValue(utils.FilterNumeric(m.numberTextInput.Value()))
	var cmd tea.Cmd
	m.numberTextInput, cmd = m.numberTextInput.Update(msg)

	return m, cmd
}

func (m MailBurstModel) View() string {
	ui := lipgloss.JoinVertical(lipgloss.Left, "發信數量：", m.numberTextInput.View(), "ctrl+c or q to quit")
	return ui
}

func InitialMailBurstModel() MailBurstModel {
	ti := textinput.New()
	ti.Width = 10
	ti.Placeholder = "1 ~ 99999"
	ti.Focus()
	ti.CharLimit = 5
	return MailBurstModel{
		numberTextInput: ti,
	}
}
