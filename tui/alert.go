// 訊息框顯示信件發送結果 最後導向到信件欄位輸入可以重複寄信
package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

// AlertModel 用於顯示發信結果的提示框
//   - Msg: 主訊息
//   - CloseMsg: 關閉提示
type AlertModel struct {
	Msg      string
	CloseMsg string
}

// Init 初始化 AlertModel
func (m AlertModel) Init() tea.Cmd {
	return nil
}

// Update 處理鍵盤事件並回到上一畫面
func (m AlertModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		// clear screen and go back to previous screen
		switch key {
		case "esc":
			// 嘗試從新的 ComposeModel 開始，如果沒有則回到舊的 MailFieldsModel
			if composeModel, ok := viper.Get("compose-model").(ComposeModel); ok {
				return composeModel, tea.ClearScreen
			}
			// 備用：回到舊的 MailFieldsModel
			anyModelInterface := viper.Get("mail-fields-model")
			mailFieldsModel, ok := anyModelInterface.(MailFieldsModel)
			if !ok {
				log.Fatalf("unexpected type: %T", anyModelInterface)
			}
			return mailFieldsModel, tea.ClearScreen
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View 渲染提示框內容
func (m AlertModel) View() string {
	dec := getAlertBuilder(m.Msg, m.CloseMsg)
	return dec.String()
}

func initAlertModel(msg string) AlertModel {
	return AlertModel{Msg: msg, CloseMsg: "[Esc] to close"}
}
