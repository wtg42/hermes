package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomEmail(t *testing.T) {
	domains := []string{"a.com", "b.org"}
	email1 := RandomEmail(domains)
	email2 := RandomEmail(domains)
	assert.NotEqual(t, email1, email2) // 應具有隨機性

	parts := strings.Split(email1, "@")
	assert.Len(t, parts, 2)
	assert.Contains(t, domains, parts[1])
	assert.NotEmpty(t, parts[0])
}

func TestRandomInt(t *testing.T) {
	seen := map[int]bool{}
	for i := 0; i < 50; i++ {
		v := RandomInt(10)
		assert.True(t, v >= 0 && v < 10)
		seen[v] = true
	}
	assert.Greater(t, len(seen), 1) // 至少出現過兩個不同值
	assert.Equal(t, 0, RandomInt(1))
}

func TestRandomString(t *testing.T) {
	s1 := RandomString(8)
	s2 := RandomString(8)
	assert.Len(t, s1, 8)
	assert.Len(t, s2, 8)
	assert.NotEqual(t, s1, s2)
	assert.Equal(t, "", RandomString(0))
	for _, c := range s1 {
		assert.True(t, strings.ContainsRune(letters, c))
	}
}
