package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterNumeric(t *testing.T) {
	assert.Equal(t, "12345", FilterNumeric("a1b2c3d4e5"))
	assert.Equal(t, "", FilterNumeric("abc"))
	assert.Equal(t, "2468", FilterNumeric("\n2@4#6$8"))
}

func TestValidateEmails(t *testing.T) {
	// 測試多個有效和無效 email
	valid, invalid := ValidateEmails("foo@example.com, bar@example.org, bad@, \t baz@domain")
	assert.Equal(t, []string{"foo@example.com", "bar@example.org"}, valid)
	assert.Equal(t, []string{"bad@", "baz@domain"}, invalid)

	// 測試空字串
	valid, invalid = ValidateEmails("")
	assert.Empty(t, valid)
	assert.Equal(t, []string{""}, invalid)

	// 測試單一有效 email
	valid, invalid = ValidateEmails("user@domain.com")
	assert.Equal(t, []string{"user@domain.com"}, valid)
	assert.Empty(t, invalid)

	// 測試多個 email 包含空格
	valid, invalid = ValidateEmails("  user1@test.com  ,  user2@test.com  ,  invalid@  ")
	assert.Equal(t, []string{"user1@test.com", "user2@test.com"}, valid)
	assert.Equal(t, []string{"invalid@"}, invalid)

	// 測試標準 email 格式（當前正則表達式支援的格式）
	valid, invalid = ValidateEmails("john@example.com, jane@example.com, invalid@")
	assert.Equal(t, []string{"john@example.com", "jane@example.com"}, valid)
	assert.Equal(t, []string{"invalid@"}, invalid)
}
