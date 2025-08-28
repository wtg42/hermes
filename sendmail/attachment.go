package sendmail

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path"

	"github.com/spf13/viper"
	"github.com/wtg42/hermes/utils"
)

// 附件結構
type Attachment struct {
	FileName    string
	FilePath    string
	ContentType string
	Encoding    string
	EncodedFile string
}

// Create a new Attachment from viper configuration.
// Returns true when attachment is available.
func (a *Attachment) NewAttachment() (bool, error) {
	// From viper db
	// 使用 viper 資料庫取得用戶的輸入設定郵件
	mailFields := viper.GetStringMap("mailField")

	// 先檢查是否有需要處理附件
	filePath, ok := mailFields["attachment"].(string)
	if !ok || filePath == "" {
		log.Println("mailField.attachment is invalid, filePath will be set to empty.")
		return false, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to open the file: %w", err)
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return false, fmt.Errorf("failed to read the file: %w", err)
	}
	encodedFile := base64.StdEncoding.EncodeToString(fileData)

	// 有副檔名先用附檔名偵測 無附檔名才檢測檔案內容
	mimeType := mime.TypeByExtension(path.Ext(filePath))
	if len(mimeType) == 0 {
		mimeType, err = utils.GetMIMEType(filePath)
		if err != nil {
			return false, fmt.Errorf("failed to get MIME type: %w", err)
		}
	}

	a.FilePath = filePath
	a.FileName = path.Base(filePath)
	a.ContentType = mimeType
	a.Encoding = "base64"
	a.EncodedFile = encodedFile

	return true, nil
}
