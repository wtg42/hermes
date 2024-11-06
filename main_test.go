package main

import (
	"testing"
)

// 測試 drawLogo 函數
func TestDrawLogo(t *testing.T) {
	// 測試輸入參數（需要確保這些是有效的測試數據，或模擬合適的數據）
	iPath := imagePath("imgs/gopher_img.png")
	fPath := fontPath("fonts/RobotoMono-Regular.ttf")

	// 調用 drawLogo 函數
	asciiArt, _ := drawLogo(iPath, fPath)

	// 驗證返回的值是否為非空字符串
	if asciiArt == "" {
		t.Errorf("Expected non-empty ASCII art, got an empty string")
	}
}

func TestDrawWithEmptyPaths(t *testing.T) {
	// 測試 iPath 為空的情況
	iPath := imagePath("")
	fPath := fontPath("/valid/font/path.ttf")

	_, err := drawLogo(iPath, fPath)
	if err == nil || err.Error() != "image path cannot be empty" {
		t.Errorf("Expected error 'image path cannot be empty', got %v", err)
	}

	// 測試 fPath 為空的情況
	iPath = imagePath("/valid/image/path.jpg")
	fPath = fontPath("")

	_, err = drawLogo(iPath, fPath)
	if err == nil || err.Error() != "font path cannot be empty" {
		t.Errorf("Expected error 'font path cannot be empty', got %v", err)
	}
}
