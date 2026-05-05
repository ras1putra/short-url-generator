package slug

import (
	"crypto/rand"
	"strings"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var reservedSlugs = []string{
	"api",
	"health",
	"login",
	"register",
	"dashboard",
	"_next",
}

func Generate(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	for i := range b {
		b[i] = charset[b[i]%62]
	}
	return string(b)
}

func IsReserved(slug string) bool {
	lower := strings.ToLower(slug)
	for _, r := range reservedSlugs {
		if lower == r {
			return true
		}
	}
	return false
}