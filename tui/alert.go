package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type AlertModel struct {
	Msg      string
	CloseMsg string
}

func (m AlertModel) Init() tea.Cmd {
	return nil
}

func (m AlertModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		// clear screen and go back to previous screen
		switch key {
		case "esc":
			return InitialAppModel(), tea.ClearScreen
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m AlertModel) View() string {
	dec := getAlertBuilder(m.Msg, m.CloseMsg)
	return dec.String()
}

func initAlertModel(msg string) AlertModel {
	return AlertModel{Msg: msg, CloseMsg: "[Esc] to close"}
}
