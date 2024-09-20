package tui

import (
	"fmt"
	"hermes/sendmail"
	"hermes/utils"
	"log"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionStatus int

const (
	quantityInput sessionStatus = iota
	hostInput
	portInput
)

type MailBurstModel struct {
	session         sessionStatus
	viewport        viewport.Model
	numberTextInput textinput.Model
	hostTextInput   textinput.Model
	portTextInput   textinput.Model
}

func (m MailBurstModel) Init() tea.Cmd {
	return nil
}

func (m MailBurstModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch m.session {
	case quantityInput:
		if !m.numberTextInput.Focused() {
			cmd = m.numberTextInput.Focus()
			cmds = append(cmds, cmd)
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "tab":
				m.session = hostInput
				m.numberTextInput.Blur()
			}
		}
		// 限制用戶只能輸入數字 並且會排除 0
		m.numberTextInput.SetValue(utils.FilterNumeric(m.numberTextInput.Value()))
		m.numberTextInput, cmd = m.numberTextInput.Update(msg)
		cmds = append(cmds, cmd)
	case hostInput:
		if !m.hostTextInput.Focused() {
			cmd = m.hostTextInput.Focus()
			cmds = append(cmds, cmd)
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "tab":
				m.session = portInput
				m.hostTextInput.Blur()
				return m, nil
			case "shift+tab":
				m.session = quantityInput
				m.hostTextInput.Blur()
				return m, nil
			}
		}
		m.hostTextInput, cmd = m.hostTextInput.Update(msg)
		cmds = append(cmds, cmd)
	case portInput:
		if !m.portTextInput.Focused() {
			cmd = m.portTextInput.Focus()
			cmds = append(cmds, cmd)
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "shift+tab":
				m.session = hostInput
				m.portTextInput.Blur()
				return m, nil
			}
		}
		m.portTextInput, cmd = m.portTextInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Enter 直接發送
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			quantity, err := strconv.ParseInt(m.numberTextInput.Value(), 10, 64)
			if err != nil || quantity <= 0 {
				fmt.Printf("Failed to convert quantity to int: %+v or quantity <= 0: \n", err)
				log.Fatalf("Failed to convert quantity to int: %+v or quantity <= 0: \n", err)
			}
			host := m.hostTextInput.Value()
			port := m.portTextInput.Value()
			sendmail.BurstModeSendMail(int(quantity), host, port)
			return m, tea.Quit
		}
	}

	// 這邊很重要 你必須對 viewport 更新上下滾動的效果才會生效
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	// 這邊重新繪製 viewport 內的 input 內容
	m.updateViewportUI()

	return m, tea.Batch(cmds...)
}

func (m MailBurstModel) View() string {
	ui := lipgloss.JoinVertical(lipgloss.Left, m.viewport.View(), "\n  ↑/↓: Navigate • Tab/Shift+Tab: Switch Focus • q: Quit\n")
	return ui
}

// 初始化 Model 跟內容
func InitialMailBurstModel() *MailBurstModel {
	// 數量
	ti := textinput.New()
	ti.Focus()
	ti.Width = 10
	ti.Placeholder = "1 ~ 99999"
	ti.CharLimit = 5

	// Host
	tiHost := textinput.New()
	tiHost.Width = 10
	tiHost.Placeholder = "localhost"
	tiHost.CharLimit = 20

	// Port
	tiPort := textinput.New()
	tiPort.Width = 10
	tiPort.Placeholder = "1025"
	tiPort.CharLimit = 5

	var b strings.Builder

	// 簡單排版
	b.WriteString(
		"發信數量: \n" +
			ti.View() +
			"\n\n" +
			"主機: \n" +
			tiHost.View() +
			"\n\n" +
			"Port: \n" +
			tiPort.View() +
			"\n",
	)

	vp := viewport.New(50, 5)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)
	vp.SetContent(b.String())

	return &MailBurstModel{
		session:         quantityInput,
		viewport:        vp,
		numberTextInput: ti,
		hostTextInput:   tiHost,
		portTextInput:   tiPort,
	}
}

func (m *MailBurstModel) updateViewportUI() {
	m.viewport.SetContent(
		"發信數量: \n" +
			m.numberTextInput.View() +
			"\n\n" +
			"主機: \n" +
			m.hostTextInput.View() +
			"\n\n" +
			"Port: \n" +
			m.portTextInput.View() + "\n",
	)
}
