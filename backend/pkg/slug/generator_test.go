package slug

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate_Length(t *testing.T) {
	s := Generate(6)
	assert.Len(t, s, 6)
}

func TestGenerate_DifferentLengths(t *testing.T) {
	for _, length := range []int{1, 5, 10, 20} {
		s := Generate(length)
		assert.Len(t, s, length)
	}
}

func TestGenerate_Charset(t *testing.T) {
	s := Generate(1000)
	for _, c := range s {
		assert.True(t, (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'))
	}
}

func TestGenerate_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		s := Generate(6)
		assert.False(t, seen[s], "generated duplicate slug: %s", s)
		seen[s] = true
	}
}

func TestIsReserved_True(t *testing.T) {
	reserved := []string{"api", "health", "login", "register", "dashboard", "_next"}
	for _, slug := range reserved {
		assert.True(t, IsReserved(slug), "expected %s to be reserved", slug)
	}
}

func TestIsReserved_CaseInsensitive(t *testing.T) {
	assert.True(t, IsReserved("API"))
	assert.True(t, IsReserved("Login"))
	assert.True(t, IsReserved("DASHBOARD"))
}

func TestIsReserved_False(t *testing.T) {
	notReserved := []string{"myslug", "abc123", "hello", "random", "xyz"}
	for _, slug := range notReserved {
		assert.False(t, IsReserved(slug), "expected %s to not be reserved", slug)
	}
}