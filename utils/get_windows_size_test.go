package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"
)

// 非 terminal 環境下應該要失敗且長寬回傳 0
func TestGetWindowSizeFail(t *testing.T) {
	// 先設定 log output 到自訂的變數
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	width, height, err := GetWindowSize()

	if err == nil {
		t.Errorf("expected error, got nil")
	}

	if width != 0 {
		t.Errorf("expected width 0, got %d", width)
	}

	if height != 0 {
		t.Errorf("expected height 0, got %d", height)
	}

	// Check log output for specific error message
	expectedLogMessage := fmt.Sprintf("Error getting terminal size: %v", err)
	if !bytes.Contains(logBuffer.Bytes(), []byte(expectedLogMessage)) {
		t.Errorf("Expected log message to contain '%s', but got '%s'", expectedLogMessage, logBuffer.String())
	}
}
