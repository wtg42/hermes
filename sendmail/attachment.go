package sendmail

import (
	"encoding/base64"
	"io"
	"log"
	"mime"
	"os"
	"path"

	"github.com/spf13/viper"
)

// 附件結構
type Attachment struct {
	FileName    string
	FilePath    string
	ContentType string
	Encoding    string
	EncodedFile string
}

// Create a new Attachment
func (a *Attachment) NewAttachment() bool {
	// From viper db
	// 使用 viper 資料庫取得用戶的輸入設定郵件
	mailFields := viper.GetStringMap("mailField")

	// 先檢查是否有需要處理附件
	filePath, ok := mailFields["attachment"].(string)
	if !ok || filePath == "" {
		log.Println("mailField.attachment is invalid, filePath will be set to empty.")
		return false
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open the file:%+v", err)
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read the file:%+v", err)
	}
	encodedFile := base64.StdEncoding.EncodeToString(fileData)

	a.FilePath = filePath
	a.FileName = path.Base(filePath)
	a.ContentType = mime.TypeByExtension(path.Ext(filePath))
	a.Encoding = "base64"
	a.EncodedFile = encodedFile

	return true
}
