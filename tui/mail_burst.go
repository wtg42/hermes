package tui

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/wtg42/hermes/sendmail"
	"github.com/wtg42/hermes/utils"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionStatus int

// 給游標判斷 focus 用的
const (
	quantityInput sessionStatus = iota
	hostInput
	portInput
	receiverDomainInput
)

// 畫面元件相關
// MailBurstModel 爆發式發信畫面模型
//   - session: 當前輸入階段
//   - viewport: 視窗顯示內容
//   - numberTextInput 等: 使用者輸入欄位
type MailBurstModel struct {
	session                 sessionStatus
	viewport                viewport.Model
	numberTextInput         textinput.Model
	hostTextInput           textinput.Model
	portTextInput           textinput.Model
	receiverDomainTextInput textinput.Model // 要接受的信件網域名
}

// Init 初始化 MailBurstModel
func (m MailBurstModel) Init() tea.Cmd {
	return nil
}

// Update 處理使用者輸入並更新畫面
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
			case "ctrl+c":
				return m, tea.Quit
			case "tab":
				m.session = receiverDomainInput
				m.portTextInput.Blur()
				return m, nil
			case "shift+tab":
				m.session = hostInput
				m.portTextInput.Blur()
				return m, nil
			}
		}
		m.portTextInput, cmd = m.portTextInput.Update(msg)
		cmds = append(cmds, cmd)
	case receiverDomainInput:
		if !m.receiverDomainTextInput.Focused() {
			cmd = m.receiverDomainTextInput.Focus()
			cmds = append(cmds, cmd)
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "shift+tab":
				m.session = portInput
				m.receiverDomainTextInput.Blur()
				return m, nil
			}
		}
		m.receiverDomainTextInput, cmd = m.receiverDomainTextInput.Update(msg)
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
			rcDomain := strings.Split(m.receiverDomainTextInput.Value(), ",")
			sendmail.BurstModeSendMail(int(quantity), host, port, rcDomain)
			return m, tea.Quit
		}
	}

	{
		// Disable some keybindings in viewport that affect input
		newKeyMap := viewport.DefaultKeyMap()
		newKeyMap.Down = key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down"))
		newKeyMap.Up = key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "down"))
		newKeyMap.PageUp = key.NewBinding()
		newKeyMap.PageDown = key.NewBinding()
		newKeyMap.HalfPageDown = key.NewBinding()
		newKeyMap.HalfPageUp = key.NewBinding()
		m.viewport.KeyMap = newKeyMap
	}
	// ⚠️ 這邊很重要 你必須對 viewport 更新上下滾動的效果才會生效
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	// 這邊重新繪製 viewport 內的 input 內容
	m.updateViewportUI()

	return m, tea.Batch(cmds...)
}

// View 渲染爆發式發信畫面
func (m MailBurstModel) View() string {
	ui := lipgloss.JoinVertical(lipgloss.Left, m.viewport.View(), "\n  ↑/↓: Navigate • Tab/Shift+Tab: Switch Focus • q: Quit\n")
	return ui
}

// InitialMailBurstModel 初始化 MailBurstModel 與預設欄位
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

	// receiverDomain
	tiReceiverDomain := textinput.New()
	tiReceiverDomain.Width = 10
	tiReceiverDomain.Placeholder = "The Domain Name of the Receiver, separated by a comma"

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
			"\n\n" +
			"Receiver Domain: \n" +
			tiReceiverDomain.View() +
			"\n",
	)

	vp := viewport.New(50, 10)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)
	vp.SetContent(b.String())

	return &MailBurstModel{
		session:                 quantityInput,
		viewport:                vp,
		numberTextInput:         ti,
		hostTextInput:           tiHost,
		portTextInput:           tiPort,
		receiverDomainTextInput: tiReceiverDomain,
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
			m.portTextInput.View() +
			"\n\n" +
			"Receiver Domain: \n" +
			m.receiverDomainTextInput.View() +
			"\n",
	)
}
