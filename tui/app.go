// 寄信程式介面 收集用戶輸入在傳給 SMTP client 發送
package tui

import (
	"go-go-power-mail/utils"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	errMsg error
)

// 主畫面 Model
type AppModel struct {
	MailFields []textinput.Model // 用戶輸入的 SMTP IP
	Focused    int               // 當前焦點的位置
	comfirm    bool              // 用戶最後確認
	err        error
}

// 樣式集合宣告
var (
	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DC851C")).Align(lipgloss.Left)
	enterButtonStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#FF4D94")).
				Foreground(lipgloss.Color("#FFFFFF")). // 這個顏色好像沒有顯示出來
				Padding(0, 2)
	cancelButtonStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#878B7D")).
				Foreground(lipgloss.Color("#FFFFFF")). // 這個顏色好像沒有顯示出來
				Padding(0, 2)
	// testStyle = lipgloss.NewStyle().
	// 		BorderStyle(lipgloss.NormalBorder()).
	// 		BorderForeground(lipgloss.Color("63"))
)

func InitialAppModel() AppModel {

	// AppModel.MailFields 數量初始化
	m := AppModel{
		MailFields: make([]textinput.Model, 5),
		comfirm:    false,
	}

	//
	for i := range m.MailFields {
		t := textinput.New()
		t.Cursor.Blink = true

		switch i {
		case 0:
			t.Placeholder = "From"
			t.CharLimit = 256
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "To"
			t.CharLimit = 0
		case 2:
			t.Placeholder = "Subject"
			t.CharLimit = 256
		case 3:
			t.Placeholder = "Contents"
			t.CharLimit = 0
		case 4:
			t.Placeholder = "Host"
			t.CharLimit = 64
		}

		m.MailFields[i] = t
	}

	return m
}

// 這個並不會被自動呼叫，因為他不是初始化的 model 你需要自行呼叫
func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c": // 這個畫面不要用字元跳出 因為使用者要輸入
			return m, tea.Quit

		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			// Cycle indexes
			if (s == "tab" || s == "down") && m.Focused < len(m.MailFields)-1 {
				m.Focused++
			}

			if (s == "shift+tab" || s == "up") && m.Focused > 0 {
				m.Focused--
			}

			// 樣式的更新
			cmds := make([]tea.Cmd, len(m.MailFields))
			for i := 0; i <= len(m.MailFields)-1; i++ {
				if i == m.Focused {
					// Set focused state
					cmds[i] = m.MailFields[i].Focus()
					m.MailFields[i].PromptStyle = focusedStyle
					m.MailFields[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.MailFields[i].Blur()
				m.MailFields[i].PromptStyle = lipgloss.NewStyle()
				m.MailFields[i].TextStyle = lipgloss.NewStyle()
			}

			return m, tea.Batch(cmds...)

		case "enter", "esc":
			s := msg.String()
			if s == "enter" {
				log.Println("TODO: 跳出 dialog 最後確認")
				m.comfirm = true
				return m, nil
			}
			if s == "esc" {
				for i := range m.MailFields {
					m.MailFields[i].SetValue("")
				}
				return m, nil
			}

		case "y", "n":
			log.Println("TODO: 用戶 Enter 進入到下一個階段，整理用戶資訊")
			return m, nil
		}
	case tea.WindowSizeMsg:
		log.Printf("width: %d, height: %d\n", msg.Width, msg.Height)
		return m, nil
	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	// 訊息內如更新
	cmds := make([]tea.Cmd, len(m.MailFields))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.MailFields {
		m.MailFields[i], cmds[i] = m.MailFields[i].Update(msg)
	}
	return m, cmd
}

func (m AppModel) View() string {

	utils.GetWindowSize()
	if m.comfirm {
		doc := getDialogBuilder()
		return doc.String()
	}

	var b strings.Builder

	// labels
	labels := []string{"寄件者: \n", "收件者: \n", "主旨: \n", "內容: \n", "信件主機: \n"}

	for i := range m.MailFields {
		b.WriteString(labels[i])
		b.WriteString(m.MailFields[i].View())

		if i < len(m.MailFields)-1 {
			b.WriteRune('\n')
			b.WriteRune('\n')
		}
	}

	// 排版換行
	b.WriteString("\n\n")

	buttons := []string{"確定[enter]", "取消[esc]"}
	for i := range buttons {
		if i == 0 {
			setStyleString := enterButtonStyle.Render(buttons[i])
			b.WriteString(setStyleString + "  ")
		} else {
			setStyleString := cancelButtonStyle.Render(buttons[i])
			b.WriteString(setStyleString + "  ")
		}
	}

	// 排版換行
	b.WriteString("\n")

	return b.String()
}

// 產生 dialog layout
func getDialogBuilder() strings.Builder {
	width, height := utils.GetWindowSize()
	doc := strings.Builder{}

	{
		var subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
		dialogBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

		buttonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			Padding(0, 3).
			MarginTop(1)

		// 之後可以支援 mouse event 可以加上底線效果
		activeButtonStyle := buttonStyle.
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#F25D94")).
			MarginRight(2)

		okButton := activeButtonStyle.Render("Yes[y]")
		cancelButton := buttonStyle.Render("No[n]")

		question := lipgloss.
			NewStyle().
			Width(50).
			Align(lipgloss.Center).
			Render("確定送出嗎?")

		buttons := lipgloss.JoinHorizontal(lipgloss.Right, okButton, cancelButton)
		ui := lipgloss.JoinVertical(lipgloss.Center, question, buttons)

		dialog := lipgloss.Place(width, height,
			lipgloss.Center, lipgloss.Center,
			dialogBoxStyle.Render(ui),
			// lipgloss.WithWhitespaceChars(""),
			lipgloss.WithWhitespaceForeground(subtle),
		)

		doc.WriteString(dialog + "\n\n")
	}
	return doc
}
