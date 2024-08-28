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
func (a *Attachment) NewAttachment() {
	// From viper db
	// 使用 viper 資料庫取得用戶的輸入設定郵件
	mailFields := viper.GetStringMap("mailField")
	filePath := mailFields["attachment"].(string)

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln("Error opening file:", err)
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	encodedFile := base64.StdEncoding.EncodeToString(fileData)

	a.FilePath = filePath
	a.FileName = path.Base(filePath)
	a.ContentType = mime.TypeByExtension(path.Ext(filePath))
	a.Encoding = "base64"
	a.EncodedFile = encodedFile
}
