package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/TheZoraiz/ascii-image-converter/aic_package"
	"github.com/spf13/viper"
	"github.com/wtg42/hermes/cmd"
	"github.com/wtg42/hermes/utils"
)

type imagePath string

func (i imagePath) String() string {
	return string(i)
}
func (i imagePath) IsEmpty() bool {
	return i.String() == ""
}

type fontPath string

func (f fontPath) String() string {
	return string(f)
}

func (f fontPath) IsEmpty() bool {
	return f.String() == ""
}

const (
	iPath imagePath = "imgs/gopher_img.png"
	fPath fontPath  = "fonts/RobotoMono-Regular.ttf"
)

func main() {
	{
		/* tea 已經實作了 log 套件的功能 */
		f, err := tea.LogToFile(os.TempDir()+"/tea_debug.log", "tea-debug")
		if err != nil {
			log.Fatalln(err)
		}
		defer f.Close()
	}

	// fetch user cmd and check if help was displayed
	shouldDrawLogo := cmd.Execute()

	// generate logo only if not showing help
	if shouldDrawLogo {
		gopherImg, err := drawLogo(iPath, fPath)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("%v\n", gopherImg)
	}

	// debug
	userInputCmd := viper.Get("userInputCmd")
	log.Println("the cmd that user inputed is => ", userInputCmd)
}

// Will draw a big image of a gopher
func drawLogo(iPath imagePath, fPath fontPath) (string, error) {
	// Can't not be empty.
	if iPath.IsEmpty() {
		return "", fmt.Errorf("image path cannot be empty")
	}

	if fPath.IsEmpty() {
		return "", fmt.Errorf("font path cannot be empty")
	}

	// 設定圖片位置
	filePath, err := utils.ExtractFile(iPath.String())
	if err != nil {
		return "", err
	}

	fontPath, err := utils.ExtractFile(fPath.String())
	if err != nil {
		return "", err
	}

	flags := aic_package.DefaultFlags()

	// This part is optional.
	// You can directly pass default flags variable to aic_package.Convert() if you wish.
	flags.Width = 70
	flags.Colored = supportsLogoColor()
	flags.Braille = true
	flags.Threshold = 1
	flags.FontFilePath = fontPath

	// Note: For environments where a terminal isn't available (such as web servers),
	// you MUST specify atleast one of flags.Width, flags.Height or flags.Dimensions
	// Conversion for an image
	asciiArt, err := aic_package.Convert(filePath, flags)
	if err != nil {
		return "", err
	}

	return asciiArt, nil
}

func supportsLogoColor() bool {
	term := strings.ToLower(os.Getenv("TERM"))
	colorTerm := strings.ToLower(os.Getenv("COLORTERM"))

	return strings.Contains(term, "256color") ||
		strings.Contains(colorTerm, "truecolor") ||
		strings.Contains(colorTerm, "24bit")
}
