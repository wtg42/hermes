// 過濾輸入類相關工具
package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// FilterNumeric 過濾輸入，只保留數字字元
//   - input: 原始字串
func FilterNumeric(input string) string {
	var b strings.Builder
	for _, r := range input {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// SplitEmails 將逗號分隔的 email 字串分割成陣列，去除前後空格
// 返回非空的 email 地址列表
func SplitEmails(userInput string) []string {
	if userInput == "" {
		return []string{}
	}

	emails := strings.Split(userInput, ",")
	var result []string
	for _, email := range emails {
		trimmed := strings.TrimSpace(email)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// ValidateEmails 驗證逗號分隔的 email 字串，返回有效的 email 列表和無效的 email 列表
func ValidateEmails(userInput string) ([]string, []string) {
	emailPattern := `(?i)^\s*\"?[a-zA-Z0-9._%+-]+\"?\s*<\s*[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\s*>$|^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	emails := strings.Split(userInput, ",")

	validEmails := []string{}
	invalidEmails := []string{}
	re := regexp.MustCompile(emailPattern)

	for _, email := range emails {
		trimmedEmail := strings.TrimSpace(email) // 去除前後空格
		if re.MatchString(trimmedEmail) {
			validEmails = append(validEmails, trimmedEmail)
		} else {
			invalidEmails = append(invalidEmails, trimmedEmail)
		}
	}

	return validEmails, invalidEmails
}
