package main

import (
	"fmt"
	"go-go-power-mail/cmd"
	"log"
	"os"

	"github.com/TheZoraiz/ascii-image-converter/aic_package"
	"github.com/spf13/viper"
)

// 定義用戶可以執行的指令
const (
	cmdStartTUI       = "start-tui"
	cmdDirectSendMail = "directSendMail"
)

func main() {
	// init log setting
	file, err := os.OpenFile("/var/tmp/debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)

	drawLogo()

	// fetch user cmd
	cmd.Execute()
	userInputCmd := viper.Get("userInputCmd")
	fmt.Println("==>", userInputCmd)

	// 依照選擇指令
	switch userInputCmd {
	case cmdStartTUI:
		// TODO: 實現 TUI 啟動邏輯
	case cmdDirectSendMail:
		// sendmail.DirectSendMail()
	default:
		fmt.Printf("我不知道你想要幹嘛?: %s\n", userInputCmd)
	}
}

// Will draw a big image of a gopher
func drawLogo() {
	// If file is in current directory. This can also be a URL to an image or gif.
	filePath := "./imgs/gopher_img.png"

	flags := aic_package.DefaultFlags()

	// This part is optional.
	// You can directly pass default flags variable to aic_package.Convert() if you wish.
	// There are more flags, but these are the ones shown for demonstration
	// flags.Dimensions = []int{100}
	flags.Width = 70
	flags.Colored = true
	flags.CustomMap = " .-=+#@"
	flags.FontFilePath = "./fonts/RobotoMono-Regular.ttf" // If file is in current directory
	// flags.SaveBackgroundColor = [4]int{50, 50, 50, 100}

	// Note: For environments where a terminal isn't available (such as web servers), you MUST
	// specify atleast one of flags.Width, flags.Height or flags.Dimensions

	// Conversion for an image
	asciiArt, err := aic_package.Convert(filePath, flags)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%v\n", asciiArt)
}
