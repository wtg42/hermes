package utils

import (
	"log"
	"os"

	"golang.org/x/term"
)

// GetWindowSize 取得目前終端機視窗尺寸
//   - 回傳 width 與 height
func GetWindowSize() (int, int, error) {
	fd := int(os.Stdin.Fd())
	width, height, err := term.GetSize(fd)
	if err != nil {
		log.Println("Error getting terminal size:", err)
		return 0, 0, err
	}

	return width, height, nil
}
