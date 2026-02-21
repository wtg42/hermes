// 統一撰寫頁面 TUI
// 整合 Header 欄位、Composer 內文、Preview 預覽於單一畫面
// 左側分割為 Header panel（上）和 Composer panel（下），右側為 Preview panel
package tui

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	"github.com/wtg42/hermes/sendmail"
	"github.com/wtg42/hermes/utils"
)

// ComposeModel 統一撰寫畫面的模型
type ComposeModel struct {
	// Header fields
	mailFields    []textinput.Model // 7 個欄位：From, To, Cc, Bcc, Subject, Host, Port
	focusedField  int               // 當前焦點的 textinput 索引 (0~6)

	// Composer
	composer textarea.Model

	// Preview
	preview viewport.Model

	// State
	activePanel int // 0 = header panel, 1 = composer panel
	width       int
	height      int

	// Filepicker overlay
	showFilePicker bool
	filepicker     filepicker.Model
	selectedFile   string

	// 發信狀態
	sending bool
	err     error

	// Esc 計數（連按兩次退出）
	escCount int
}

// 樣式集合
var (
	focusedBorderColor = lipgloss.Color("#DC851C") // 橘色
	normalBorderColor  = lipgloss.Color("62")      // 灰色
	focusedHeaderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DC851C")).
		Align(lipgloss.Left)
)

// InitialComposeModel 初始化 ComposeModel
func InitialComposeModel() ComposeModel {
	w, h, err := utils.GetWindowSize()
	if err != nil {
		log.Fatalf("Error getting terminal size: %v", err)
	}

	// 初始化 mailFields（複用 MailFieldsModel 的邏輯）
	mailFields := make([]textinput.Model, 7)
	for i := range mailFields {
		t := textinput.New()
		t.Cursor.Blink = true

		switch i {
		case 0:
			t.Placeholder = "From"
			t.CharLimit = 256
			t.Focus()
			t.PromptStyle = focusedHeaderStyle
			t.TextStyle = focusedHeaderStyle
		case 1:
			t.Placeholder = "To"
			t.CharLimit = 512
		case 2:
			t.Placeholder = "Cc"
			t.CharLimit = 512
		case 3:
			t.Placeholder = "Bcc"
			t.CharLimit = 512
		case 4:
			t.Placeholder = "Subject"
			t.CharLimit = 256
		case 5:
			t.Placeholder = "Host"
			t.CharLimit = 64
		case 6:
			t.Placeholder = "default is 25"
			t.CharLimit = 6
		}

		mailFields[i] = t
	}

	// 初始化 composer textarea
	composer := textarea.New()
	composer.Placeholder = "Compose your email here..."
	composer.SetHeight(10)

	// 初始化 preview viewport
	rightWidth := w / 2
	previewHeight := h - 1 // viewport.Height 是含邊框的總高度，只扣狀態列

	preview := viewport.New(rightWidth, previewHeight)
	preview.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(normalBorderColor).
		PaddingRight(2)
	preview.KeyMap = viewport.KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
	}

	// 初始化 filepicker
	fp := filepicker.New()
	fp.AllowedTypes = []string{} // 允許所有檔案類型
	fp.ShowHidden = false
	fp.CurrentDirectory, _ = os.UserHomeDir()

	m := ComposeModel{
		mailFields:     mailFields,
		focusedField:   0,
		composer:       composer,
		preview:        preview,
		activePanel:    0, // 預設焦點在 Header panel
		width:          w,
		height:         h,
		showFilePicker: false,
		filepicker:     fp,
		selectedFile:   "",
		sending:        false,
		escCount:       0,
	}

	return m
}

// Init 初始化命令
func (m ComposeModel) Init() tea.Cmd {
	return m.filepicker.Init()
}

// Update 處理鍵盤事件與模型更新
func (m ComposeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// 更新各 panel 的寬高
		rightWidth := m.width / 2
		previewHeight := m.height - 1 // viewport.Height 是含邊框的總高度，只扣狀態列

		m.preview.Width = rightWidth
		m.preview.Height = previewHeight
		return m, nil

	case sendMailProcess:
		// 發信完成，顯示結果
		var warning string
		if msg.err != nil {
			warning = "😩 " + msg.err.Error()
		} else {
			warning = "🎉 信件傳送成功"
		}

		// 保存當前狀態以便返回
		viper.Set("compose-model", m)

		return initAlertModel(warning), tea.ClearScreen
	}

	// 處理 Filepicker Overlay 的消息（需在 tea.KeyMsg 之前處理）
	if m.showFilePicker {
		isFilePickerReadDirMsg := fmt.Sprintf("%T", msg)
		if isFilePickerReadDirMsg == "filepicker.readDirMsg" {
			var fpCmd tea.Cmd
			m.filepicker, fpCmd = m.filepicker.Update(msg)

			// 檢查是否選擇了檔案
			if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
				m.selectedFile = path
				m.showFilePicker = false
				return m, nil
			}

			return m, fpCmd
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 處理 Filepicker Overlay 的按鍵
		if m.showFilePicker {
			switch msg.String() {
			case "esc":
				// 關閉 Overlay
				m.showFilePicker = false
				return m, nil

			default:
				// 交給 filepicker 處理
				var fpCmd tea.Cmd
				m.filepicker, fpCmd = m.filepicker.Update(msg)
				return m, fpCmd
			}
		}

		// 全域快捷鍵
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "ctrl+s":
			// 觸發發信
			return m.handleSend()

		case "esc":
			m.escCount++
			if m.escCount >= 2 {
				return m, tea.Quit
			}
			// 清空所有欄位
			for i := range m.mailFields {
				m.mailFields[i].SetValue("")
			}
			m.composer.SetValue("")
			m.preview.SetContent("")
			m.escCount = 0
			return m, nil

		case "ctrl+a":
			// 觸發附件選取 Overlay
			m.showFilePicker = true
			return m, nil

		case "ctrl+j":
			// 從 Header 切換到 Composer
			if m.activePanel == 0 {
				m.activePanel = 1
				// Blur all textinputs
				for i := range m.mailFields {
					m.mailFields[i].Blur()
					m.mailFields[i].PromptStyle = lipgloss.NewStyle()
					m.mailFields[i].TextStyle = lipgloss.NewStyle()
				}
				m.composer.Focus()
				return m, nil
			}

		case "ctrl+k":
			// 從 Composer 切換到 Header
			if m.activePanel == 1 {
				m.activePanel = 0
				m.composer.Blur()
				m.mailFields[m.focusedField].Focus()
				m.mailFields[m.focusedField].PromptStyle = focusedHeaderStyle
				m.mailFields[m.focusedField].TextStyle = focusedHeaderStyle
				return m, nil
			}
		}

		// Panel 特定的按鍵處理
		if m.activePanel == 0 {
			return m.handleHeaderKeys(msg)
		} else if m.activePanel == 1 {
			return m.handleComposerKeys(msg)
		}
	}

	// 預設：更新 composer 並同步 preview
	if m.activePanel == 1 {
		var cmd tea.Cmd
		m.composer, cmd = m.composer.Update(msg)
		// 同步 preview 內容
		m.preview.SetContent(m.composer.Value())
		return m, cmd
	}

	return m, nil
}

// handleHeaderKeys 處理 Header panel 的按鍵
func (m ComposeModel) handleHeaderKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		// 在 Header 欄位間循環
		m.focusedField = (m.focusedField + 1) % len(m.mailFields)
		cmds := make([]tea.Cmd, len(m.mailFields))
		for i := range m.mailFields {
			if i == m.focusedField {
				cmds[i] = m.mailFields[i].Focus()
				m.mailFields[i].PromptStyle = focusedHeaderStyle
				m.mailFields[i].TextStyle = focusedHeaderStyle
			} else {
				m.mailFields[i].Blur()
				m.mailFields[i].PromptStyle = lipgloss.NewStyle()
				m.mailFields[i].TextStyle = lipgloss.NewStyle()
			}
		}
		return m, tea.Batch(cmds...)

	case "shift+tab":
		// 反向循環
		m.focusedField = (m.focusedField - 1 + len(m.mailFields)) % len(m.mailFields)
		cmds := make([]tea.Cmd, len(m.mailFields))
		for i := range m.mailFields {
			if i == m.focusedField {
				cmds[i] = m.mailFields[i].Focus()
				m.mailFields[i].PromptStyle = focusedHeaderStyle
				m.mailFields[i].TextStyle = focusedHeaderStyle
			} else {
				m.mailFields[i].Blur()
				m.mailFields[i].PromptStyle = lipgloss.NewStyle()
				m.mailFields[i].TextStyle = lipgloss.NewStyle()
			}
		}
		return m, tea.Batch(cmds...)

	default:
		// 交給當前焦點的 textinput 處理
		var cmd tea.Cmd
		m.mailFields[m.focusedField], cmd = m.mailFields[m.focusedField].Update(msg)
		return m, cmd
	}
}

// handleComposerKeys 處理 Composer panel 的按鍵
func (m ComposeModel) handleComposerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+h":
		// 填入 HTML 範本
		m.composer.SetValue(htmlTemplate)
		m.preview.SetContent(htmlTemplate)
		return m, nil

	case "ctrl+t":
		// 填入 Plain Text 範本
		m.composer.SetValue(textTemplate)
		m.preview.SetContent(textTemplate)
		return m, nil

	case "ctrl+e":
		// 填入 EML 範本
		m.composer.SetValue(emlTemplate)
		m.preview.SetContent(emlTemplate)
		return m, nil

	default:
		// 交給 textarea 處理
		var cmd tea.Cmd
		m.composer, cmd = m.composer.Update(msg)
		// 同步 preview 內容
		m.preview.SetContent(m.composer.Value())
		return m, cmd
	}
}

// handleSend 觸發發信流程
func (m ComposeModel) handleSend() (tea.Model, tea.Cmd) {
	m.sending = true // 設定發信中狀態
	// 保存所有欄位值到 viper
	viper.Set("mailField.from", m.mailFields[0].Value())
	viper.Set("mailField.to", m.mailFields[1].Value())
	viper.Set("mailField.cc", m.mailFields[2].Value())
	viper.Set("mailField.bcc", m.mailFields[3].Value())
	viper.Set("mailField.subject", m.mailFields[4].Value())
	viper.Set("mailField.host", m.mailFields[5].Value())
	viper.Set("mailField.port", m.mailFields[6].Value())
	viper.Set("mailField.contents", m.composer.Value())
	viper.Set("mailField.selectedFile", m.selectedFile)

	// 呼叫發信函數（非同步）
	return m, sendMailWithChannel()
}

// sendMailWithChannel 非同步發信（複用舊設計邏輯）
func sendMailWithChannel() tea.Cmd {
	return func() tea.Msg {
		success, err := sendmail.SendMailWithMultipart("mailField")
		return sendMailProcess{
			result: success,
			err:    err,
		}
	}
}

// View 渲染統一撰寫畫面
func (m ComposeModel) View() string {
	leftWidth := m.width / 2
	// 可用內容高度 = terminal 高度 - 狀態列(1) - 兩個 panel 邊框(4)
	available := m.height - 5
	if available < 5 {
		available = 5
	}
	headerHeight := available * 2 / 5
	composerHeight := available - headerHeight // 取剩餘，避免整數捨去產生誤差

	// 渲染 Header panel
	headerContent := m.renderHeaderPanel(leftWidth, headerHeight)

	// 渲染 Composer panel
	composerContent := m.renderComposerPanel(leftWidth, composerHeight)

	// 左側版面：Header + Composer（垂直堆疊）
	leftPane := lipgloss.JoinVertical(
		lipgloss.Left,
		headerContent,
		composerContent,
	)

	// 右側版面：Preview
	rightPane := m.preview.View()

	// 左右分割
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		rightPane,
	)

	// 底部狀態列
	statusBar := m.renderStatusBar()

	// 如果顯示 Filepicker Overlay
	if m.showFilePicker {
		fpHeight := m.height - 4
		fpContent := lipgloss.NewStyle().
			Width(leftWidth).
			Height(fpHeight).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(focusedBorderColor).
			Render(m.filepicker.View())

		// 將 Overlay 置中於 Composer 區域
		fpOverlay := lipgloss.Place(
			m.width,
			m.height-2,
			lipgloss.Center,
			lipgloss.Center,
			fpContent,
		)

		return lipgloss.JoinVertical(
			lipgloss.Top,
			fpOverlay,
			statusBar,
		)
	}

	// 正常版面：mainContent + statusBar
	return lipgloss.JoinVertical(
		lipgloss.Top,
		mainContent,
		statusBar,
	)
}

// renderHeaderPanel 渲染 Header panel
func (m ComposeModel) renderHeaderPanel(width, height int) string {
	borderColor := normalBorderColor
	if m.activePanel == 0 {
		borderColor = focusedBorderColor
	}

	headerStyle := lipgloss.NewStyle().
		Width(width - 4).
		Height(height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		PaddingRight(1).
		PaddingLeft(1)

	// 組合 7 個欄位
	var fields []string
	for _, field := range m.mailFields {
		fields = append(fields, fmt.Sprintf("%s", field.View()))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, fields...)
	return headerStyle.Render(content)
}

// renderComposerPanel 渲染 Composer panel
func (m ComposeModel) renderComposerPanel(width, height int) string {
	borderColor := normalBorderColor
	if m.activePanel == 1 {
		borderColor = focusedBorderColor
	}

	composerStyle := lipgloss.NewStyle().
		Width(width - 4).
		Height(height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		PaddingRight(1).
		PaddingLeft(1)

	m.composer.SetWidth(width - 6)
	m.composer.SetHeight(height - 2)

	return composerStyle.Render(m.composer.View())
}

// renderStatusBar 渲染底部狀態列
func (m ComposeModel) renderStatusBar() string {
	// 若正在發信，顯示等待提示
	if m.sending {
		return lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("214")).
			Render("⏳ Sending... Please wait")
	}

	// 根據當前 panel 動態顯示相關快捷鍵
	var panelHint string
	if m.activePanel == 0 {
		panelHint = "  [Tab] Next Field  [Ctrl+J] Compose"
	} else {
		panelHint = "  [Ctrl+K] Header"
	}

	// 快捷鍵提示
	shortcuts := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("[Ctrl+S] Send  [Ctrl+A] Attach  [Esc×2] Quit" + panelHint)

	// SMTP 連線狀態
	host := m.mailFields[5].Value()
	port := m.mailFields[6].Value()
	if port == "" {
		port = "25"
	}

	connStatus := ""
	if host != "" {
		connStatus = fmt.Sprintf("Connected to %s:%s • TLS active", host, port)
	}

	connStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(connStatus)

	// 組合狀態列
	statusBar := lipgloss.JoinHorizontal(
		lipgloss.Center,
		shortcuts,
		"  ",
		connStyle,
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(statusBar)
}
