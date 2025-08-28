package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestMailBurstModel_UpdateSessionAndFilterNumeric 確認 Tab 會切換 session 並過濾非數字
func TestMailBurstModel_UpdateSessionAndFilterNumeric(t *testing.T) {
	m := InitialMailBurstModel()

	// 模擬輸入字母與數字，Tab 切換到下一欄位後以 q 結束
	input := strings.NewReader("a1b2\tq")
	p := tea.NewProgram(m, tea.WithInput(input), tea.WithoutSignalHandler(), tea.WithoutRenderer())

	finalModel, err := p.Run()
	assert.NoError(t, err)

	model, ok := finalModel.(MailBurstModel)
	if assert.True(t, ok, "final model should be MailBurstModel") {
		// Tab 後應切換到 hostInput
		assert.Equal(t, hostInput, model.session)
		// 非數字應被過濾，只留下 "12"
		assert.Equal(t, "12", model.numberTextInput.Value())
	}
}
