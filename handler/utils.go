package handler

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateRandomCode gera um código numérico aleatório de N dígitos.
func GenerateRandomCode(length int) (string, error) {
	const digits = "0123456789"
	code := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", fmt.Errorf("erro ao gerar código aleatório: %v", err)
		}
		code[i] = digits[num.Int64()]
	}

	return string(code), nil
}
