// 過濾輸入類相關工具
package utils

import (
	"strings"
	"unicode"
)

// 過濾輸入，只保留數字字符
func FilterNumeric(input string) string {
	var b strings.Builder
	for _, r := range input {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
