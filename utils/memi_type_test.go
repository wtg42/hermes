package utils

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMIMEType(t *testing.T) {
	// text file
	txt, err := os.CreateTemp("", "sample*.txt")
	assert.NoError(t, err)
	// 寫入超過 512 bytes 的純文字，避免被當成二進位資料
	_, err = txt.WriteString(strings.Repeat("a", 600))
	assert.NoError(t, err)
	txt.Close()
	t.Cleanup(func() { os.Remove(txt.Name()) })

	mime, err := GetMIMEType(txt.Name())
	assert.NoError(t, err)
	assert.Equal(t, "text/plain; charset=utf-8", mime)

	// png file
	pngFile, err := os.CreateTemp("", "img*.png")
	assert.NoError(t, err)
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	err = png.Encode(pngFile, img)
	assert.NoError(t, err)
	pngFile.Close()
	t.Cleanup(func() { os.Remove(pngFile.Name()) })

	mime, err = GetMIMEType(pngFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, "image/png", mime)

	// pdf file
	pdfFile, err := os.CreateTemp("", "doc*.pdf")
	assert.NoError(t, err)
	_, err = pdfFile.Write([]byte("%PDF-1.4\n"))
	assert.NoError(t, err)
	pdfFile.Close()
	t.Cleanup(func() { os.Remove(pdfFile.Name()) })

	mime, err = GetMIMEType(pdfFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, "application/pdf", mime)
}
