// menu 顯示所有功能的畫面
package tui

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuModel struct {
	choices []string
	cursor  int
	// Actually, it is similar to true/false, but it doesn't take up memory.
	selected map[int]struct{}
	done     bool // 表示用戶選擇完成
}

// menu 原始字串 要渲染特效請用這個才不會有重複再次渲染問題
var menuOptions = []string{"自訂郵件內容發送", "Burst Mode", "使用 eml 發送", "Quit"}

var (
	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	// The style of the currently selected option
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#DC851C"))
)

func initialMenuModel() menuModel {
	return menuModel{
		choices:  styledChoices(),
		selected: make(map[int]struct{}),
	}
}

func (m menuModel) Init() tea.Cmd {

	return nil
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// 用戶模式決定後結束並回傳選擇
		case "enter", " ":
			// Quit
			if m.cursor == 3 {
				m.selected[m.cursor] = struct{}{}
				m.done = true // 設置為完成
				return m, tea.Quit
			}

			// click 進行 mail_field.go 畫面顯示
			m.selected[m.cursor] = struct{}{}
			m.done = true // 設置為完成

			// 根據選項返回相對應的 model
			// 回傳的是 model 初始化的 func
			returnUserSelectedOptionModel := func() func() (tea.Model, tea.Cmd) {
				switch m.cursor {
				case 0:
					return func() (tea.Model, tea.Cmd) {
						return InitialMailFieldsModel(), nil
					}
				case 1:
					return func() (tea.Model, tea.Cmd) {
						return InitialMailBurstModel(), nil
					}
				case 2:
					return func() (tea.Model, tea.Cmd) {
						// Filepicker is a special module.
						// You need to invoke its Init() first.
						// then return the "cmd"
						emlModel := InitialEmlModel()
						cmd := emlModel.filepicker.Init()
						return emlModel, cmd
					}
				default:
					return func() (tea.Model, tea.Cmd) {
						return m, nil
					}
				}
			}()

			return returnUserSelectedOptionModel()
		}
	}

	// Update cursor style
	for i := range m.choices {
		if i == m.cursor {
			m.choices[i] = cursorStyle.Render(menuOptions[i])
		} else {
			m.choices[i] = normalStyle.Render(menuOptions[i]) // Revert to normal style
		}
	}

	return m, nil
}

func (m menuModel) View() string {
	s := "Choose an option:\n\n"

	for i, choice := range m.choices {

		// 使用兩個空格 因為 👉 很寬
		cursor := "  "
		if m.cursor == i {
			cursor = "👉"
		}

		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\nPress q to quit.\n"

	return s
}

// TUI 程式畫面的起點
// 希望在這裡可以回傳 smtp 需要的資訊
func StartMenu() (int, bool, tea.Model) {
	m := initialMenuModel()
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		log.Fatalf("發生錯誤：%v", err)
	}

	// 結束時候的判斷 判定用戶執行到在哪一個步驟(which model)
	switch model := finalModel.(type) {
	case menuModel:
		if model.done {
			return model.cursor, true, model
		}
		return -1, false, nil
	case MailFieldsModel:
		// 處理 AppModel 的情況
		// 這裡可能需要根據您的需求返回適當的值
		return 0, true, model
	default:
		fmt.Printf("未知的模型類型：%T\n", finalModel)
		return -1, false, nil
	}
}

// menu 選項樣式渲染
func styledChoices() []string {
	styledChoices := make([]string, len(menuOptions))
	for i, choice := range menuOptions {
		// Apply normal style initially
		styledChoices[i] = normalStyle.Render(choice)
	}

	return styledChoices
}
