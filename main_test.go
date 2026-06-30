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
	asciiArt, err := drawLogo(iPath, fPath)
	if err != nil {
		t.Fatalf("Expected drawLogo to succeed, got %v", err)
	}

	// 驗證返回的值是否為非空字符串
	if asciiArt == "" {
		t.Errorf("Expected non-empty ASCII art, got an empty string")
	}
}

func TestSupportsLogoColor(t *testing.T) {
	tests := []struct {
		name      string
		term      string
		colorTerm string
		want      bool
	}{
		{
			name: "dumb terminal",
			term: "dumb",
			want: false,
		},
		{
			name: "256 color terminal",
			term: "xterm-256color",
			want: true,
		},
		{
			name:      "truecolor terminal",
			term:      "xterm",
			colorTerm: "truecolor",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TERM", tt.term)
			t.Setenv("COLORTERM", tt.colorTerm)

			if got := supportsLogoColor(); got != tt.want {
				t.Fatalf("supportsLogoColor() = %v, want %v", got, tt.want)
			}
		})
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
