package utils

import (
	"io"
	"os"
	"path/filepath"

	"github.com/wtg42/hermes/assets"
)

// ExtractFile 將嵌入資源寫入臨時檔並回傳路徑
//   - embedPath: 嵌入檔案在 assets 中的位置
func ExtractFile(embedPath string) (string, error) {
	data, err := assets.StaticFiles.Open(embedPath)
	if err != nil {
		return "", err
	}
	defer data.Close()

	// 留空會自動在暫時目錄產生檔案
	tempFile, err := os.CreateTemp("", filepath.Base(embedPath))
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, data)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}
