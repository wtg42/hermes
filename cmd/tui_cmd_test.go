package cmd

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/wtg42/hermes/tui"
)

// TestStartTuiCmdRunCalled 確認 Run 會呼叫 TUI
func TestStartTuiCmdRunCalled(t *testing.T) {
	root := &cobra.Command{Use: "hermes"}
	root.AddCommand(startTUICmd)

	called := false
	original := tui.StartMenu
	tui.StartMenu = func() (int, bool, tea.Model) {
		called = true
		return 0, true, nil
	}
	defer func() { tui.StartMenu = original }()

	root.SetArgs([]string{"start-tui"})
	if _, err := root.ExecuteC(); err != nil {
		t.Fatalf("執行命令應無錯誤: %v", err)
	}
	if !called {
		t.Fatalf("預期 StartMenu 被呼叫")
	}
}
