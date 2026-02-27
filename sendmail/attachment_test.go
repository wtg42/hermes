package sendmail

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

// 測試新的 NewAttachment 函數
func TestNewAttachment(t *testing.T) {
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

	// 測試新的 NewAttachment 函數（接受 filePath 參數）
	a, err := NewAttachment(tempFile.Name())
	if err != nil {
		t.Errorf("NewAttachment returned error: %v", err)
	}
	if a == nil {
		t.Errorf("Expected NewAttachment to return a non-nil Attachment")
		return
	}

	// 驗證 Attachment 的欄位是否被正確設定
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

// 測試舊的 NewAttachmentLegacy 方法（為向後相容保留）
func TestNewAttachmentLegacy(t *testing.T) {
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

	viper.Set("mailField", map[string]any{
		"attachment": tempFile.Name(),
	})

	a := Attachment{}
	result, err := a.NewAttachmentLegacy()
	if err != nil {
		t.Errorf("NewAttachmentLegacy returned error: %v", err)
	}
	// 驗證結果是否為 true
	if !result {
		t.Errorf("Expected NewAttachmentLegacy to return true, but got false")
	}

	// 驗證 Attachment 的欄位是否被正確設定
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
