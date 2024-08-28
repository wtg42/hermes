package utils

import (
	"log"
	"net/http"
	"os"
)

// GetMIMEType 用來檢測檔案內容的 MIME 類型
func GetMIMEType(filePath string) (string, error) {
	// 打開文件並讀取前 512 個字節（因為 MIME 類型檢測只需要前 512 個字節）
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error opening file:", err)
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		log.Println("Error reading file:", err)
		return "", err
	}

	// 根據文件內容檢測 MIME 類型
	mimeType := http.DetectContentType(buffer)
	log.Println("Detected MIME type:", mimeType)

	return mimeType, nil
}
