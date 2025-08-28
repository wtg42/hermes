// generate random string, int, anything you want.
package utils

import (
	"strings"
	"time"

	"math/rand"
)

const letters = "abcdefghijklmnopqrstuvwxyz"

// RandomEmail 產生隨機 Email 地址
//   - domains: 可用的網域列表
func RandomEmail(domains []string) string {
	// 創建一個隨機數生成器的實例，基於當前時間的 seed
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	// 初始化 byte slice 並填充隨機字母
	randStr := func() string {
		// 隨機長度 至少要有一個長度 不燃 Email 格式會錯誤
		rs := make([]byte, RandomInt(5)+1)
		for i := range rs {
			rs[i] = letters[r.Intn(len(letters))]
		}
		return string(rs)
	}

	var b = strings.Builder{}
	b.WriteString(randStr())
	b.WriteString("@")
	// Randomly select a domain from domains
	b.WriteString(domains[r.Intn(len(domains))])
	// b.WriteString("rd01.softnext.com")

	// 創建一個 fake email
	// 如果你產生 poper name 寄信出去好像都會自動產生 Bcc 欄位
	// fakeEmail := mail.Address{
	// Name: "", Address: b.String(),
	// }
	// return fakeEmail.String()

	return b.String()
}

// RandomInt 回傳 0 到 n-1 的隨機整數
//   - n: 上限值
func RandomInt(n int) int {
	source := rand.NewSource(time.Now().UnixNano()) // 使用當前時間納秒數作為種子
	r := rand.New(source)                           // 創建一個新的隨機數生成器
	randomNumber := r.Intn(n)                       // 生成 0 到 n-1 之間的隨機整數
	return randomNumber
}

// RandomString 隨機生成字串
//   - n: 字串長度
func RandomString(n int) string {
	var sb strings.Builder
	k := len(letters)
	for i := 0; i < n; i++ {
		c := letters[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}
