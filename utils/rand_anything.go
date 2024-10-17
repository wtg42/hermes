// generate random string, int, anything you want.
package utils

import (
	"log"
	"net/mail"
	"strings"
	"time"

	"math/rand"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// RandomEmail 產生一個隨機的 Email 地址
//   - 產生一個隨機數字作為 Email 的 local part
//   - 產生一個隨機字串作為 Email 的 domain
//   - 產生一個隨機字串作為 Email 的 Alias
//   - 產生一個 fake 的 Email  Address 實例
//   - 產生一個隨機的 Email 字串
func RandomEmail(domains []string) string {
	// 創建一個隨機數生成器的實例，基於當前時間的 seed
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	// 初始化 byte slice 並填充隨機字母
	randStr := func() string {
		// 隨機長度 至少要有一個長度 不燃 Email 格式會錯誤
		b := make([]byte, RandomInt(5)+1)
		for i := range b {
			b[i] = letters[r.Intn(len(letters))]
		}
		return string(b)
	}

	var b = strings.Builder{}
	b.WriteString(randStr())
	b.WriteString("@")
	// Randomly select a domain from domains
	b.WriteString(domains[r.Intn(len(domains))])
	// b.WriteString("rd01.softnext.com")

	// 創建一個 fake email
	fakeEmail := mail.Address{
		Name: randStr(), Address: b.String(),
	}

	log.Printf("==>%s", fakeEmail.String())
	return fakeEmail.String()
}

func RandomInt(n int) int {
	source := rand.NewSource(time.Now().UnixNano()) // 使用當前時間納秒數作為種子
	r := rand.New(source)                           // 創建一個新的隨機數生成器
	randomNumber := r.Intn(n)                       // 生成 0 到 n-1 之間的隨機整數
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
