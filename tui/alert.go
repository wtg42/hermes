package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
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
			anyModelInterface := viper.Get("app-model")
			appModel, ok := anyModelInterface.(AppModel)
			if !ok {
				log.Fatalf("unexpected type: %T", anyModelInterface)
			}
			appModel.Comfirm = false
			return appModel, tea.ClearScreen
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
