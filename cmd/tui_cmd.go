package cmd

import (
	"hermes/tui"

	"github.com/spf13/cobra"
)

var startTUICmd = &cobra.Command{
	Use:   "start-tui",
	Short: "啟動文字用戶界面（TUI）",
	Long:  `啟動應用程序的文字用戶界面（TUI）模式，提供互動式操作環境。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 這裡添加啟動 TUI 的邏輯
		tui.StartMenu()
	},
}
