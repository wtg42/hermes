// generate random string, int, anything you want.
package utils

import (
	"net/mail"
	"strings"
	"time"

	"math/rand"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomEmail() string {
	// 創建一個隨機數生成器的實例，基於當前時間的 seed
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	// 初始化 byte slice 並填充隨機字母
	randStr := func() string {
		b := make([]byte, RandomInt(5))
		for i := range b {
			b[i] = letters[r.Intn(len(letters))]
		}
		return string(b)
	}

	var b = strings.Builder{}
	b.WriteString(randStr())
	b.WriteString("@")
	b.WriteString(randStr())
	b.WriteString(".com")

	// 創建一個 fake email
	fakeEmail := mail.Address{
		Name: randStr(), Address: b.String(),
	}

	return fakeEmail.String()
}

func RandomInt(n int) int {
	source := rand.NewSource(time.Now().UnixNano()) // 使用當前時間納秒數作為種子
	r := rand.New(source)                           // 創建一個新的隨機數生成器
	randomNumber := r.Intn(n)                       // 生成 0 到 99 之間的隨機數
	return randomNumber
}

// 隨機生成字串
func RandomString(n int) string {
	var sb strings.Builder
	k := len(letters)
	for i := 0; i < n; i++ {
		c := letters[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}
