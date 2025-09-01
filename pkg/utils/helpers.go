package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// RandomHex generates a random hexadecimal string of length n
func RandomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
