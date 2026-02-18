package common

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// GeneratePIN generates a numeric PIN of specified length
func GeneratePIN(length int) string {
	if length < 4 {
		length = 4
	}
	if length > 8 {
		length = 8
	}
	numbers := "0123456789"
	result := make([]byte, length)
	maxParam := big.NewInt(int64(len(numbers)))

	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, maxParam)
		result[i] = numbers[n.Int64()]
	}

	return string(result)
}

// GenerateVoucherCode generates a random string based on type
func GenerateVoucherCode(length int, charType string) string {
	const (
		numbers = "0123456789"
		alpha   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		mixed   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	)

	var charset string
	switch charType {
	case "number":
		charset = numbers
	case "alpha":
		charset = alpha
	default:
		charset = mixed
	}

	// Remove confusing characters
	charset = strings.ReplaceAll(charset, "0", "")
	charset = strings.ReplaceAll(charset, "O", "")
	charset = strings.ReplaceAll(charset, "1", "")
	charset = strings.ReplaceAll(charset, "I", "")
	charset = strings.ReplaceAll(charset, "l", "")

	result := make([]byte, length)
	maxParam := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, maxParam)
		result[i] = charset[n.Int64()]
	}

	return string(result)
}

// RandStr generates a random string of fixed length (mixed)
func RandStr(length int) string {
	return GenerateVoucherCode(length, "mixed")
}
