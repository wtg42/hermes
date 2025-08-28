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
	valid, invalid := ValidateEmails("foo@example.com, bar@example.org, bad@, \t baz@domain")
	assert.Equal(t, []string{"foo@example.com", "bar@example.org"}, valid)
	assert.Equal(t, []string{"bad@", "baz@domain"}, invalid)

	valid, invalid = ValidateEmails("")
	assert.Empty(t, valid)
	assert.Equal(t, []string{""}, invalid)
}
