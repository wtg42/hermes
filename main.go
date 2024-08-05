package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

func main() {
	// 設置 Viper 來讀取配置文件
	viper.SetConfigName("config") // 設置配置文件名稱（不需要文件擴展名）
	viper.SetConfigType("yaml")   // 如果配置文件的擴展名不是 yaml，需要顯式設置
	viper.AddConfigPath(".")      // 設置查找配置文件的路徑

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	appName := viper.GetString("app_name")
	// 寄件者
	sender := viper.GetString("sender")

	contentBody := viper.GetString("contentBody")

	fmt.Println("your app_name string is :", appName)

	fmt.Println("contnet body is :", contentBody)

	fmt.Println("sender is :", sender)
}
