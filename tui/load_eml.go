// 使用 eml 發信的畫面
package tui

import (
	"log"
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EmlModel struct {
	status       sessionStatus
	selectedFile string
	filepicker   filepicker.Model
}

const (
	isFilepicker sessionStatus = iota
)

func (m EmlModel) Init() tea.Cmd {
	return nil
}

func (m EmlModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.status {
	case isFilepicker:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, tea.Quit
			}
			m.filepicker, cmd = m.filepicker.Update(msg)
			// Did the user select a file?
			if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
				// Get the path of the selected file.
				m.selectedFile = path
			}
			return m, cmd
		}
		// The other msg behavior still needs to update the filepicker.
		m.filepicker, cmd = m.filepicker.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m EmlModel) View() string {
	pickfileDscription := "\nPick a .eml file: "
	if m.selectedFile != "" {
		pickfileDscription = "\nSelected file: " + m.filepicker.Styles.Selected.Render(m.selectedFile)
	}

	pickfileDscription = focusedStyle.Render(pickfileDscription)
	ui := lipgloss.JoinVertical(lipgloss.Left, pickfileDscription, m.filepicker.View())

	return ui
}

// 初始化 EmlModel
func InitialEmlModel() EmlModel {
	fp := filepicker.New()
	fp.Height = 5
	fp.AllowedTypes = []string{".eml"}
	var err error
	fp.CurrentDirectory, err = os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	return EmlModel{
		selectedFile: "",
		status:       isFilepicker,
		filepicker:   fp,
	}
}
