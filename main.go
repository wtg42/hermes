package main

import (
	"fmt"
	"log"
	"os"

	"github.com/TheZoraiz/ascii-image-converter/aic_package"
	tea "github.com/charmbracelet/bubbletea"
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

	// generate logo...
	gopherImg, err := drawLogo(iPath, fPath)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%v\n", gopherImg)

	// fetch user cmd
	cmd.Execute()

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
		log.Fatalln(err)
	}

	fontPath, err := utils.ExtractFile(fPath.String())
	if err != nil {
		log.Fatalln(err)
	}

	flags := aic_package.DefaultFlags()

	// This part is optional.
	// You can directly pass default flags variable to aic_package.Convert() if you wish.
	flags.Width = 70
	flags.Colored = true
	flags.Braille = true
	flags.Threshold = 1
	flags.FontFilePath = fontPath

	// Note: For environments where a terminal isn't available (such as web servers),
	// you MUST specify atleast one of flags.Width, flags.Height or flags.Dimensions
	// Conversion for an image
	asciiArt, err := aic_package.Convert(filePath, flags)
	if err != nil {
		log.Fatalln(err)
	}

	return asciiArt, nil
}
