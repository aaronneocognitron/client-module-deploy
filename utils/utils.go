package utils

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"time"
)

func IsTimeZero(t time.Time) bool {
	return t.Equal(time.Time{})
}

func InSlice[T comparable](element T, slice []T) bool {
	for _, val := range slice {
		if val == element {
			return true
		}
	}

	return false
}

func MapKeys[T comparable, R any](object map[T]R) []T {
	keys := make([]T, len(object))

	i := 0
	for k := range object {
		keys[i] = k
		i++
	}

	return keys
}

func GenerateRandomBytes(size int) ([]byte, error) {
	bytes := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return nil, err
	}

	return bytes, nil
}

func GeneratePassword(length int) (string, error) {
	const characters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return GenerateRandomString(characters, length)
}

func GenerateEncryptionString(length int) (string, error) {
	const characters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()%?*~{}"
	return GenerateRandomString(characters, length)
}

func GenerateRandomString(characters string, length int) (string, error) {
	charactersLength := len(characters)

	randomString := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(charactersLength)))
		if err != nil {
			return "", err
		}
		randomString[i] = characters[n.Int64()]
	}

	return string(randomString), nil
}

func Pkcs7Padding(input []byte, blockSize int) []byte {
	r := len(input) % blockSize
	pl := blockSize - r
	for i := 0; i < pl; i++ {
		input = append(input, byte(pl))
	}
	return input
}

func Pkcs7UnPadding(input []byte) ([]byte, error) {
	if input == nil || len(input) == 0 {
		return nil, nil
	}

	pc := input[len(input)-1]
	pl := int(pc)
	err := CheckPkcs7PaddingIsValid(input, pl)
	if err != nil {
		return nil, err
	}
	return input[:len(input)-pl], nil
}

func CheckPkcs7PaddingIsValid(input []byte, paddingLength int) error {
	if len(input) < paddingLength {
		return fmt.Errorf("invalid bytes padding")
	}
	p := input[len(input)-(paddingLength):]
	for _, pc := range p {
		if uint(pc) != uint(len(p)) {
			return fmt.Errorf("invalid bytes padding")
		}
	}
	return nil
}
