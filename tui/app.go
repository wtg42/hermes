package tui

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
)

func init() {
	fmt.Println("tui init")
}

func Start() {
	// 初始化 tcell 螢幕
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Failed to create screen: %v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("Failed to initialize screen: %v", err)
	}
	defer screen.Fini()

	// 設定預設樣式
	defStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	screen.SetStyle(defStyle)
	screen.Clear()

	// 顯示簡單的文字
	putString(screen, 10, 5, "Hello, tcell!")

	// 刷新螢幕以顯示文字
	screen.Show()

	// 事件處理迴圈
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			// 按下 'q' 鍵退出
			if ev.Key() == tcell.KeyRune && ev.Rune() == 'q' {
				return
			}
		case *tcell.EventResize:
			screen.Sync()
		}
	}
}

// 在螢幕指定位置顯示文字
func putString(s tcell.Screen, x, y int, str string) {
	for i, r := range str {
		s.SetContent(x+i, y, r, nil, tcell.StyleDefault)
	}
}
