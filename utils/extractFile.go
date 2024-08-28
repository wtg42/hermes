package utils

import (
	"go-go-power-mail/assets"
	"io"
	"os"
	"path/filepath"
)

// extractFile 將嵌入的文件寫入臨時文件並返回路徑
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
