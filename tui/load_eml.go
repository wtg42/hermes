// 使用 eml 發信的畫面
package tui

import (
	"log"
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wtg42/hermes/mail"
)

// sessionStatus 定義 EML Model 的各種狀態
type sessionStatus int

// EmlModel 提供從 .eml 檔案載入郵件的介面
//   - selectedFile: 使用者選擇的檔案
//   - filepicker: 檔案選擇元件
//   - mailer: 郵件發送器（依賴注入）
type EmlModel struct {
	status       sessionStatus
	selectedFile string
	filepicker   filepicker.Model
	mailer       mail.Mailer
}

const (
	isFilepicker sessionStatus = iota
)

// Init 初始化 EmlModel
func (m EmlModel) Init() tea.Cmd {
	return nil
}

// Update 處理檔案選擇與鍵盤事件
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

// View 渲染檔案選擇畫面
func (m EmlModel) View() string {
	pickfileDscription := "\nPick a .eml file: "
	if m.selectedFile != "" {
		pickfileDscription = "\nSelected file: " + m.filepicker.Styles.Selected.Render(m.selectedFile)
	}

	pickfileDscription = focusedStyle.Render(pickfileDscription)
	ui := lipgloss.JoinVertical(lipgloss.Left, pickfileDscription, m.filepicker.View())

	return ui
}

// InitialEmlModel 初始化 EmlModel
// 接受 mail.Mailer 依賴，用於發送 EML 郵件
func InitialEmlModel(mailer mail.Mailer) EmlModel {
	fp := filepicker.New()
	fp.SetHeight(5)
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
		mailer:       mailer,
	}
}
