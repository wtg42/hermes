// generate random string, int, anything you want.
package utils

import (
	"time"

	"math/rand"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomString(n int) string {
	// 創建一個隨機數生成器的實例，基於當前時間的 seed
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	// 初始化 byte slice 並填充隨機字母
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}

	return string(b)
}
