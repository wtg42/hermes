package sendmail

import (
	"encoding/base64"
	"testing"
)

// 測試 encodeRFC2047 函數
func TestEncodeRFC2047(t *testing.T) {
	// 測試用的輸入字串
	input := "Hello, 世界"
	expected := "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(input)) + "?="

	// 調用待測函數
	result := encodeRFC2047(input)

	// 驗證結果是否符合預期
	if result != expected {
		t.Errorf("Expected %q, but got %q", expected, result)
	}
}
