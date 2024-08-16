package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuModel struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
	done     bool // 新增：表示選擇完成
}

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

		case "enter", " ":
			if m.cursor == len(m.choices)-1 {
				m.done = true // 設置為完成
				return m, tea.Quit
			}
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
				m.done = true // 設置為完成
				return m, nil
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

func StartMenu() (int, bool) {
	m := initialMenuModel()
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("發生錯誤：%v", err)
		os.Exit(1)
	}

	finalMenuModel := finalModel.(menuModel)
	if finalMenuModel.done {
		return finalMenuModel.cursor, true
	}
	return -1, false
}

func styledChoices() []string {

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	choices := []string{"快速發送一封文字郵件", "自訂郵件發送", "Quit"}

	styledChoices := make([]string, len(choices))
	for i, choice := range choices {
		// Apply normal style initially
		styledChoices[i] = normalStyle.Render(choice)
	}

	return styledChoices
}
