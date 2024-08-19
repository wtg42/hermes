package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuModel struct {
	choices []string
	cursor  int
	// Actually, it is similar to true/false, but it doesn't take up memory.
	selected map[int]struct{}
	done     bool // 新增：表示用戶選擇完成
}

// menu 原始字串 要渲染特效請用這個才不會有重複渲染問題
var menuOptions = []string{"快速發送一封文字郵件", "自訂郵件發送", "郵件夾檔發送", "Quit"}

var (
	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	// 這個部分是 copy 一個新的 style 但是加上刪除線效果
	strikethroughStyle = normalStyle.
				Foreground(lipgloss.Color("#9E9E9E")).
				Strikethrough(true)

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
			// 沒有實作的就擋住 不需要反應任何效果
			if m.cursor > 0 && m.cursor != len(m.choices)-1 {
				return m, nil
			}

			// Quit
			if m.cursor == len(m.choices)-1 {
				m.selected[m.cursor] = struct{}{}
				m.done = true // 設置為完成
				return m, tea.Quit
			}

			// click 進行 app.go 畫面顯示
			m.selected[m.cursor] = struct{}{}
			m.done = true // 設置為完成
			return InitialAppModel(), nil
		}
	}

	// Update cursor style
	for i := range m.choices {
		if i <= 0 || i == len(menuOptions)-1 {
			if i == m.cursor {
				m.choices[i] = cursorStyle.Render(menuOptions[i])
			} else {
				m.choices[i] = normalStyle.Render(menuOptions[i]) // Revert to normal style
			}
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
		fmt.Printf("發生錯誤：%v", err)
		os.Exit(1)
	}

	// 結束時候的判斷 判定用戶執行到在哪一個步驟(which model)
	switch model := finalModel.(type) {
	case menuModel:
		fmt.Println(model)
		if model.done {
			return model.cursor, true, model
		}
		return -1, false, nil
	case AppModel:
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
		// 目前先實作第一項功能 其餘先用刪除線表示待實作 但 Quit 要可以使用
		if i > 0 && i != len(menuOptions)-1 {
			styledChoices[i] = strikethroughStyle.Render(choice)
			continue
		}
		// Apply normal style initially
		styledChoices[i] = normalStyle.Render(choice)
	}

	return styledChoices
}
