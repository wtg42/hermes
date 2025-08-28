package sendmail

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

// 測試附件是否正常從 viper 取得
// 且 NewAttachment 結構正常建立
func TestNewAttacement(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_attachment_*.txt")
	if err != nil {
		t.Errorf("Failed to create a temporary file: %v", err)
	}
	// Remove the temporary file when the test finishes.
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte("Test fake content."))
	if err != nil {
		t.Errorf("Failed to write to temporary file: %v", err)
	}
	tempFile.Close()

	viper.Set("mailField", map[string]interface{}{
		"attachment": tempFile.Name(),
	})

	a := Attachment{}
	result, err := a.NewAttachment()
	if err != nil {
		t.Errorf("NewAttachment returned error: %v", err)
	}
	// 驗證結果是否為 true
	if !result {
		t.Errorf("Expected NewAttachment to return true, but got false")
	}

	// 5. 驗證 Attachment 的欄位是否被正確設定
	if a.FilePath != tempFile.Name() {
		t.Errorf("Expected FilePath to be %s, but got %s", tempFile.Name(), a.FilePath)
	}
	if a.FileName == "" {
		t.Errorf("Expected FileName to be set, but got empty")
	}
	if a.ContentType == "" {
		t.Errorf("Expected ContentType to be set, but got empty")
	}
	if a.Encoding != "base64" {
		t.Errorf("Expected Encoding to be base64, but got %s", a.Encoding)
	}
	if a.EncodedFile == "" {
		t.Errorf("Expected EncodedFile to be set, but got empty")
	}
}
