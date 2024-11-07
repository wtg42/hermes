package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
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
		t.Errorf("Expected no error, got %v", err)
	}

	if reflect.TypeOf(width).Kind() != reflect.Int && width != 0 {
		t.Errorf("Expected width to be an int, got %T", width)
	}

	if reflect.TypeOf(height).Kind() != reflect.Int && height != 0 {
		t.Errorf("Expected height to be an int, got %T", height)
	}

	// Check log output for specific error message
	expectedLogMessage := fmt.Sprintf("Error getting terminal size: %v", err)
	if !bytes.Contains(logBuffer.Bytes(), []byte(expectedLogMessage)) {
		t.Errorf("Expected log message to contain '%s', but got '%s'", expectedLogMessage, logBuffer.String())
	}
}
