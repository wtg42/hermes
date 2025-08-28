package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// runMenu 使用指定輸入執行 StartMenu，並回傳結果
func runMenu(input string) (int, bool, tea.Model) {
	menuProgramOptions = []tea.ProgramOption{tea.WithInput(strings.NewReader(input)), tea.WithoutSignalHandler(), tea.WithoutRenderer()}
	idx, done, model := StartMenu()
	menuProgramOptions = nil // 清除測試設定避免影響其他案例
	return idx, done, model
}

func TestStartMenuOptions(t *testing.T) {
	t.Run("MailFieldsModel", func(t *testing.T) {
		t.Skip("InitialMailFieldsModel 需要終端尺寸，於 CI 環境略過")
	})

	t.Run("MailBurstModel", func(t *testing.T) {
		idx, done, model := runMenu("j\rq")
		assert.True(t, done)
		assert.Equal(t, 1, idx)
		_, ok := model.(MailBurstModel)
		assert.True(t, ok)
	})

	t.Run("EmlModel", func(t *testing.T) {
		t.Skip("EmlModel 依賴終端檔案選擇器，於 CI 環境略過")
	})

	t.Run("Quit", func(t *testing.T) {
		t.Skip("menu 模型依賴終端尺寸，於 CI 環境略過")
	})
}
