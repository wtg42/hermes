package main

import (
	"fmt"
	"go-go-power-mail/cmd"
	"log"
	"os"
	"path/filepath"

	"github.com/TheZoraiz/ascii-image-converter/aic_package"
	"github.com/spf13/viper"
)

func main() {
	// init log setting
	file, err := os.OpenFile("/var/tmp/debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(err)
	}

	log.SetOutput(file)

	drawLogo()

	// fetch user cmd
	cmd.Execute()

	// debug
	userInputCmd := viper.Get("userInputCmd")
	log.Println("the cmd that user inputed is => ", userInputCmd)
}

// Will draw a big image of a gopher
func drawLogo() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	// 設定圖片位置
	filePath := filepath.Join(cwd, "./imgs/gopher_img.png")

	flags := aic_package.DefaultFlags()

	// This part is optional.
	// You can directly pass default flags variable to aic_package.Convert() if you wish.
	flags.Width = 70
	flags.Colored = true
	flags.CustomMap = " .-=+#@"
	flags.FontFilePath = filepath.Join(cwd, "./fonts/RobotoMono-Regular.ttf") // If file is in current directory

	// Note: For environments where a terminal isn't available (such as web servers), you MUST
	// specify atleast one of flags.Width, flags.Height or flags.Dimensions

	// Conversion for an image
	asciiArt, err := aic_package.Convert(filePath, flags)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%v\n", asciiArt)
}
