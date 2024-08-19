package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type menuModel struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
	done     bool // 新增：表示選擇完成
}

func initialMenuModel() menuModel {
	return menuModel{
		choices:  []string{"Option 1", "Option 2", "Option 3", "Quit"},
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
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m menuModel) View() string {
	s := "Choose an option:\n\n"

	for i, choice := range m.choices {

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
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
