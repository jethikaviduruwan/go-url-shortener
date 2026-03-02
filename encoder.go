package main

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateCode creates a cryptographically random short code of length n.
func GenerateCode(n int) (string, error) {
	var sb strings.Builder
	sb.Grow(n)
	max := big.NewInt(int64(len(alphabet)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		sb.WriteByte(alphabet[idx.Int64()])
	}
	return sb.String(), nil
}

// IsValidCode checks that a custom code only contains safe characters.
func IsValidCode(code string) bool {
	if len(code) < 3 || len(code) > 32 {
		return false
	}
	for _, c := range code {
		if !strings.ContainsRune(alphabet+"-_", c) {
			return false
		}
	}
	return true
}
// v3-0
// v6-1
// v10-0
// v15-0
// v21-1
