package cmd

import (
	"fmt"
	"go-go-power-mail/tui"

	"github.com/spf13/cobra"
)

var startTUICmd = &cobra.Command{
	Use:   "start-tui",
	Short: "啟動文字用戶界面（TUI）",
	Long:  `啟動應用程序的文字用戶界面（TUI）模式，提供互動式操作環境。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("啟動 TUI 模式")
		// 這裡添加啟動 TUI 的邏輯
		selectedIndex, ok := tui.StartMenu()
		if ok {
			fmt.Printf("用戶選擇了選項：%d\n", selectedIndex)
		} else {
			fmt.Println("用戶沒有選擇任何選項就退出了")
		}
	},
}
