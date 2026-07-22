package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type themePicker struct {
	themes        []string
	index         int
	originalTheme string
}

func newThemePicker(currentTheme string) *themePicker {
	themes := AvailableThemes()
	index := 0
	for i, name := range themes {
		if name == currentTheme {
			index = i
			break
		}
	}
	return &themePicker{
		themes:        themes,
		index:         index,
		originalTheme: currentTheme,
	}
}

func (p *themePicker) move(delta int) string {
	p.index = (p.index + delta + len(p.themes)) % len(p.themes)
	return p.themes[p.index]
}

func (m ComposeModel) handleThemePickerKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "down", "j":
		name := m.themePicker.move(1)
		theme, _ := ResolveTheme(name)
		m.applyTheme(theme)
	case "up", "k":
		name := m.themePicker.move(-1)
		theme, _ := ResolveTheme(name)
		m.applyTheme(theme)
	case "enter":
		m.themePicker = nil
	case "esc":
		name := m.themePicker.originalTheme
		m.themePicker = nil
		theme, _ := ResolveTheme(name)
		m.applyTheme(theme)
	}
	return m, nil
}

func (m ComposeModel) renderThemePicker() string {
	theme := m.currentTheme()
	lines := []string{"Theme Picker", ""}
	for i, name := range m.themePicker.themes {
		marker := "  "
		if i == m.themePicker.index {
			marker = "> "
		}
		line := marker + name
		if i == m.themePicker.index {
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Accent)).
				Background(lipgloss.Color(theme.Selection)).
				Bold(true).
				Render(line)
		}
		lines = append(lines, line)
	}
	lines = append(lines, "", "[↑/k] Previous  [↓/j] Next", "[Enter] Apply  [Esc] Cancel")

	return borderedPanelStyle(theme, theme.Accent).
		Width(42).
		Padding(1, 2).
		Render(strings.Join(lines, "\n") + fmt.Sprintf("\n\nCurrent: %s", m.currentTheme().Name))
}
