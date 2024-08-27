package utils

import (
	"log"
	"os"

	"golang.org/x/term"
)

// GetWindowSize retrieves the current terminal window size.
func GetWindowSize() (int, int) {
	fd := int(os.Stdin.Fd())
	width, height, err := term.GetSize(fd)
	if err != nil {
		log.Println("Error getting terminal size:", err)
	}
	log.Println("====")
	log.Printf("Width: %d, Height: %d\n", width, height)
	log.Println("====")

	return width, height
}
