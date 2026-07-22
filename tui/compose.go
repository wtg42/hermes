// 統一撰寫頁面 TUI
// 整合 Header 欄位、Composer 內文、Preview 預覽於單一畫面
// 左側分割為 Header panel（上）和 Composer panel（下），右側為 Preview panel
package tui

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/viper"
	"github.com/wtg42/hermes/mail"
	"github.com/wtg42/hermes/sendmail"
	"github.com/wtg42/hermes/utils"
)

// ComposeModel 統一撰寫畫面的模型
type ComposeModel struct {
	// Header fields
	mailFields   []textinput.Model // 7 個欄位：From, To, Cc, Bcc, Subject, Host, Port
	focusedField int               // 當前焦點的 textinput 索引 (0~6)

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
	sending             bool
	err                 error
	pendingConfirmation *sendConfirmation

	// Prefix command 狀態
	prefix commandPrefix

	// 郵件發送器（依賴注入）
	mailer mail.Mailer
}

type sendConfirmation struct {
	compose mail.MailCompose
	reasons []string
	input   textinput.Model
	err     string
}

// 樣式集合
var (
	focusedPanelBorderColor = lipgloss.Color("#DC851C")
)

// sendMailProcess 發信完成訊息
type sendMailProcess struct {
	result bool
	err    error
}

// Email content templates
const (
	htmlTemplate = `<html>
<head>
    <title>Email Template</title>
</head>
<body>
    <h1>Hello!</h1>
    <p>This is an HTML email template.</p>
    <p>Best regards,<br>Your Name</p>
</body>
</html>`

	textTemplate = `Hello,

This is a plain text email template.

Best regards,
Your Name`

	emlTemplate = `Return-Path: <sender@example.com>
Received: by smtp.example.com id 123456; Mon, 1 Jan 2024 12:00:00 +0000
Date: Mon, 1 Jan 2024 12:00:00 +0000
From: Sender Name <sender@example.com>
To: Recipient Name <recipient@example.com>
Subject: Test Email
Content-Type: text/plain; charset=UTF-8

Hello,

This is a sample EML email content.

Best regards,
Sender Name`
)

// InitialComposeModel 初始化 ComposeModel
// 接受 mail.Mailer 依賴，用於發送郵件
func InitialComposeModel(mailer mail.Mailer) ComposeModel {
	w, h, err := utils.GetWindowSize()
	if err != nil {
		log.Fatalf("Error getting terminal size: %v", err)
	}

	// 初始化 mailFields（複用 MailFieldsModel 的邏輯）
	mailFields := make([]textinput.Model, 7)
	for i := range mailFields {
		t := textinput.New()

		switch i {
		case 0:
			t.Placeholder = "FROM"
			t.CharLimit = 256
			t.Focus()
		case 1:
			t.Placeholder = "TO"
			t.CharLimit = 512
		case 2:
			t.Placeholder = "CC"
			t.CharLimit = 512
		case 3:
			t.Placeholder = "BCC"
			t.CharLimit = 512
		case 4:
			t.Placeholder = "SUBJECT"
			t.CharLimit = 256
		case 5:
			t.Placeholder = "HOST"
			t.CharLimit = 64
		case 6:
			t.Placeholder = "DEFAULT IS 25"
			t.CharLimit = 6
		}

		mailFields[i] = t
	}

	// 初始化 composer textarea
	composer := textarea.New()
	composer.Placeholder = "Compose your email here..."
	composer.SetHeight(10)

	// 初始化 preview viewport
	_, rightWidth := splitPaneWidths(w)
	previewHeight := contentPaneHeight(h) // viewport.Height 是含邊框的總高度

	preview := viewport.New(
		viewport.WithWidth(previewContentWidth(rightWidth)),
		viewport.WithHeight(previewHeight),
	)
	preview.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1)
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
		prefix:         newCommandPrefix(),
		mailer:         mailer,
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
		_, rightWidth := splitPaneWidths(m.width)
		previewHeight := contentPaneHeight(m.height) // viewport.Height 是含邊框的總高度

		m.preview.SetWidth(previewContentWidth(rightWidth))
		m.preview.SetHeight(previewHeight)
		return m, nil

	case sendMailProcess:
		// 發信完成，顯示結果
		m.sending = false
		m.err = msg.err
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

	// 處理 Filepicker Overlay 的消息（需在 tea.KeyPressMsg 之前處理）
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
	case tea.KeyPressMsg:
		if m.pendingConfirmation != nil {
			return m.handleConfirmationKey(msg)
		}

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

		var command commandID
		var handled bool
		m.prefix, command, handled = m.prefix.Update(msg)
		if handled {
			return m.handleCommand(command)
		}

		// 全域快捷鍵
		switch msg.String() {
		case "ctrl+c":
			return m, nil

		case "ctrl+s":
			// 觸發發信
			return m.handleSend()

		case "esc":
			return m, nil

		case "ctrl+j":
			// 從 Header 切換到 Composer
			if m.activePanel == 0 {
				m.activePanel = 1
				// Blur all textinputs
				for i := range m.mailFields {
					m.mailFields[i].Blur()
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
func (m ComposeModel) handleHeaderKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		// 在 Header 欄位間循環
		m.focusedField = (m.focusedField + 1) % len(m.mailFields)
		cmds := make([]tea.Cmd, len(m.mailFields))
		for i := range m.mailFields {
			if i == m.focusedField {
				cmds[i] = m.mailFields[i].Focus()
			} else {
				m.mailFields[i].Blur()
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
			} else {
				m.mailFields[i].Blur()
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
func (m ComposeModel) handleComposerKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// 交給 textarea 處理
	var cmd tea.Cmd
	m.composer, cmd = m.composer.Update(msg)
	// 同步 preview 內容
	m.preview.SetContent(m.composer.Value())
	return m, cmd
}

func (m ComposeModel) handleCommand(command commandID) (tea.Model, tea.Cmd) {
	switch command {
	case commandNone, commandTemplates, commandHelp:
		return m, nil
	case commandAttach:
		m.showFilePicker = true
		return m, nil
	case commandClear:
		if !m.isDirty() {
			return m, nil
		}
		m.prefix.mode = commandModeConfirmClear
		return m, nil
	case commandConfirmClear:
		m.clearCompose()
		return m, nil
	case commandQuit:
		if !m.isDirty() {
			return m, tea.Quit
		}
		m.prefix.mode = commandModeConfirmQuit
		return m, nil
	case commandConfirmQuit:
		return m, tea.Quit
	case commandTemplateHTML:
		m.applyTemplate(htmlTemplate)
		return m, nil
	case commandTemplateText:
		m.applyTemplate(textTemplate)
		return m, nil
	case commandTemplateEML:
		m.applyTemplate(emlTemplate)
		return m, nil
	default:
		return m, nil
	}
}

func (m ComposeModel) isDirty() bool {
	for i := range m.mailFields {
		if m.mailFields[i].Value() != "" {
			return true
		}
	}
	return m.composer.Value() != "" || m.selectedFile != ""
}

func (m *ComposeModel) clearCompose() {
	for i := range m.mailFields {
		m.mailFields[i].SetValue("")
	}
	m.composer.SetValue("")
	m.preview.SetContent("")
	m.selectedFile = ""
}

func (m *ComposeModel) applyTemplate(template string) {
	m.composer.SetValue(template)
	m.preview.SetContent(template)
}

// handleSend 觸發發信流程
func (m ComposeModel) handleSend() (tea.Model, tea.Cmd) {
	compose := m.currentCompose()
	assessment, err := sendmail.AssessSingleSendSafety(sendmail.SingleSendSafetyInput{
		From: compose.From,
		To:   compose.To,
		CC:   compose.CC,
		BCC:  compose.BCC,
		Port: compose.Port,
	})
	if err != nil {
		m.err = err
		m.sending = false
		return m, nil
	}

	if assessment.RequiresConfirmation() {
		input := textinput.New()
		input.Placeholder = "SEND"
		input.CharLimit = 16
		input.Focus()
		m.err = nil
		m.pendingConfirmation = &sendConfirmation{
			compose: compose,
			reasons: append([]string(nil), assessment.Reasons...),
			input:   input,
		}
		return m, nil
	}

	return m.startSend(compose)
}

func (m ComposeModel) currentCompose() mail.MailCompose {
	to := utils.SplitEmails(m.mailFields[1].Value())
	cc := utils.SplitEmails(m.mailFields[2].Value())
	bcc := utils.SplitEmails(m.mailFields[3].Value())
	port := m.mailFields[6].Value()
	if port == "" {
		port = "25"
	}

	return mail.MailCompose{
		From:       m.mailFields[0].Value(),
		To:         to,
		CC:         cc,
		BCC:        bcc,
		Subject:    m.mailFields[4].Value(),
		Body:       m.composer.Value(),
		Attachment: m.selectedFile,
		Host:       m.mailFields[5].Value(),
		Port:       port,
	}
}

func (m ComposeModel) startSend(compose mail.MailCompose) (tea.Model, tea.Cmd) {
	m.sending = true
	m.err = nil

	// 保存當前狀態以便返回（UI 狀態管理）
	viper.Set("compose-model", m)

	// 呼叫發信函數（非同步）
	return m, m.sendMailWithChannel(compose)
}

func (m ComposeModel) handleConfirmationKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// Compose 的退出統一由 Ctrl+X、q 處理；安全確認狀態也不得繞過 prefix flow。
		return m, nil

	case "esc":
		m.pendingConfirmation = nil
		m.err = nil
		return m, nil

	case "enter":
		if m.pendingConfirmation.input.Value() != "SEND" {
			m.pendingConfirmation.err = "Confirmation must be exactly SEND"
			return m, nil
		}

		current := m.currentCompose()
		if !mailComposeSnapshotsEqual(current, m.pendingConfirmation.compose) {
			m.pendingConfirmation = nil
			m.err = fmt.Errorf("message changed; review safety and confirm again")
			return m, nil
		}

		compose := m.pendingConfirmation.compose
		m.pendingConfirmation = nil
		return m.startSend(compose)

	default:
		var cmd tea.Cmd
		m.pendingConfirmation.input, cmd = m.pendingConfirmation.input.Update(msg)
		m.pendingConfirmation.err = ""
		return m, cmd
	}
}

func mailComposeSnapshotsEqual(left, right mail.MailCompose) bool {
	return left.From == right.From &&
		slices.Equal(left.To, right.To) &&
		slices.Equal(left.CC, right.CC) &&
		slices.Equal(left.BCC, right.BCC) &&
		left.Subject == right.Subject &&
		left.Body == right.Body &&
		left.Attachment == right.Attachment &&
		slices.Equal(left.Attachments, right.Attachments) &&
		left.Host == right.Host &&
		left.Port == right.Port
}

// sendMailWithChannel 非同步發信
func (m ComposeModel) sendMailWithChannel(compose mail.MailCompose) tea.Cmd {
	return func() tea.Msg {
		err := m.mailer.Send(compose)
		return sendMailProcess{
			result: err == nil,
			err:    err,
		}
	}
}

// View 渲染統一撰寫畫面
func (m ComposeModel) View() tea.View {
	leftWidth, rightWidth := splitPaneWidths(m.width)
	paneHeight := contentPaneHeight(m.height)
	headerHeight, composerHeight := splitLeftPaneHeights(paneHeight)

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
	m.preview.SetWidth(previewContentWidth(rightWidth))
	m.preview.SetHeight(paneHeight)
	rightPane := m.preview.View()

	// 左右分割
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		rightPane,
	)

	// 底部狀態列
	statusBar := m.renderStatusBar()
	if m.pendingConfirmation != nil {
		overlayHeight := m.height - 2
		if overlayHeight < 1 {
			overlayHeight = 1
		}
		confirmationOverlay := lipgloss.Place(
			m.width,
			overlayHeight,
			lipgloss.Center,
			lipgloss.Center,
			m.renderSafetyConfirmation(),
		)
		content := lipgloss.JoinVertical(
			lipgloss.Top,
			confirmationOverlay,
			statusBar,
		)
		view := tea.NewView(content)
		view.AltScreen = true
		return view
	}

	if m.prefix.mode == commandModeHelp {
		helpContent := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(1, 2).
			Render(renderCommandHelp())
		helpOverlay := lipgloss.Place(
			m.width,
			m.height-1,
			lipgloss.Center,
			lipgloss.Center,
			helpContent,
		)
		content := lipgloss.JoinVertical(lipgloss.Top, helpOverlay, statusBar)
		view := tea.NewView(content)
		view.AltScreen = true
		return view
	}

	// 如果顯示 Filepicker Overlay
	if m.showFilePicker {
		fpHeight := m.height - 4
		fpContent := lipgloss.NewStyle().
			Width(leftWidth).
			Height(fpHeight).
			BorderStyle(lipgloss.RoundedBorder()).
			Render(m.filepicker.View())

		// 將 Overlay 置中於 Composer 區域
		fpOverlay := lipgloss.Place(
			m.width,
			m.height-2,
			lipgloss.Center,
			lipgloss.Center,
			fpContent,
		)

		content := lipgloss.JoinVertical(
			lipgloss.Top,
			fpOverlay,
			statusBar,
		)
		view := tea.NewView(content)
		view.AltScreen = true
		return view
	}

	// 正常版面：mainContent + statusBar
	content := lipgloss.JoinVertical(
		lipgloss.Top,
		mainContent,
		statusBar,
	)
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

func (m ComposeModel) renderSafetyConfirmation() string {
	lines := []string{
		"⚠ Outside safe send whitelist",
		"",
	}
	for _, reason := range m.pendingConfirmation.reasons {
		lines = append(lines, "- "+reason)
	}
	lines = append(lines,
		"",
		"Type SEND to confirm",
		m.pendingConfirmation.input.View(),
		"[Esc] Cancel",
	)
	if m.pendingConfirmation.err != "" {
		lines = append(lines, "", m.pendingConfirmation.err)
	}

	width := m.width - 8
	if width < 20 {
		width = 20
	}
	return lipgloss.NewStyle().
		Width(width).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Padding(1, 2).
		Render(strings.Join(lines, "\n"))
}

// renderHeaderPanel 渲染 Header panel
func (m ComposeModel) renderHeaderPanel(width, height int) string {
	headerStyle := lipgloss.NewStyle().
		Width(width).
		Height(height+2).
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1).
		MarginBottom(0)

	if m.activePanel == 0 {
		headerStyle = headerStyle.BorderForeground(focusedPanelBorderColor)
	}

	inputWidth := headerInputWidth(width)

	// 組合 7 個欄位
	var fields []string
	for i := range m.mailFields {
		m.mailFields[i].SetWidth(inputWidth)
		fields = append(fields, fmt.Sprintf("%s", m.mailFields[i].View()))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, fields...)
	return headerStyle.Render(content)
}

func previewContentWidth(paneWidth int) int {
	if paneWidth < 1 {
		return 1
	}

	return paneWidth
}

func splitPaneWidths(totalWidth int) (int, int) {
	if totalWidth < 2 {
		return 1, 1
	}

	leftWidth := (totalWidth + 1) / 2
	rightWidth := totalWidth - leftWidth

	if rightWidth < 1 {
		rightWidth = 1
		leftWidth = totalWidth - rightWidth
		if leftWidth < 1 {
			leftWidth = 1
		}
	}

	return leftWidth, rightWidth
}

func contentPaneHeight(totalHeight int) int {
	paneHeight := totalHeight - 1 // 預留 1 行給狀態列
	if paneHeight < 1 {
		return 1
	}

	return paneHeight
}

func splitLeftPaneHeights(paneHeight int) (int, int) {
	contentHeight := paneHeight - 4 // 兩個 panel 邊框(4)
	if contentHeight < 5 {
		contentHeight = 5
	}

	headerHeight := contentHeight * 2 / 5
	if headerHeight < 7 {
		headerHeight = 7
	}
	if headerHeight > contentHeight-1 {
		headerHeight = contentHeight - 1
	}
	if headerHeight < 1 {
		headerHeight = 1
	}

	composerHeight := contentHeight - headerHeight
	if composerHeight < 1 {
		composerHeight = 1
	}

	return headerHeight, composerHeight
}

func headerInputWidth(paneWidth int) int {
	inputWidth := paneWidth - 8
	if inputWidth < 6 {
		return 6
	}

	return inputWidth
}

// renderComposerPanel 渲染 Composer panel
func (m ComposeModel) renderComposerPanel(width, height int) string {
	composerStyle := lipgloss.NewStyle().
		Width(width).
		Height(height+2).
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1)

	if m.activePanel == 1 {
		composerStyle = composerStyle.BorderForeground(focusedPanelBorderColor)
	}

	m.composer.SetWidth(width - 4)
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
	if m.err != nil {
		return lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("196")).
			Render("⚠ " + m.err.Error())
	}

	var commandStatus string
	switch m.prefix.mode {
	case commandModeRoot, commandModeTemplate:
		commandStatus = renderCommandHUD(m.prefix.mode)
	case commandModeConfirmClear:
		commandStatus = "CLEAR  Compose content will be lost.  [C] Confirm  [Esc] Cancel"
	case commandModeConfirmQuit:
		commandStatus = "QUIT  Unsaved compose will be lost.  [Q] Confirm  [Esc] Cancel"
	case commandModeHelp:
		commandStatus = "HELP  [Esc] Close"
	}
	if commandStatus != "" {
		return m.renderCenteredStatus(commandStatus, "214")
	}
	if m.prefix.notice != "" {
		return m.renderCenteredStatus(m.prefix.notice, "214")
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
		Render("[Ctrl+S] Send  [Ctrl+X] Commands" + panelHint)

	// SMTP 連線狀態
	host := m.mailFields[5].Value()
	port := m.mailFields[6].Value()
	if port == "" {
		port = "25"
	}

	connStatus := ""
	if host != "" {
		connStatus = fmt.Sprintf("SMTP target %s:%s", host, port)
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

func (m ComposeModel) renderCenteredStatus(content, colorName string) string {
	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color(colorName)).
		Render(content)
}
