package utils

import (
	"embed"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wtg42/hermes/assets"
)

//go:embed testdata/dummy.txt
var testFiles embed.FS

func TestExtractFile(t *testing.T) {
	// 替換原始靜態資源為測試用資料
	original := assets.StaticFiles
	assets.StaticFiles = testFiles
	defer func() { assets.StaticFiles = original }()

	// 呼叫被測函式
	path, err := ExtractFile("testdata/dummy.txt")
	assert.NoError(t, err)
	t.Cleanup(func() { os.Remove(path) })

	// 讀取檔案內容並驗證
	data, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, "hello world\n", string(data))
}
