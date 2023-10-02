package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

type Encryptor struct {
	key          string
	salt         string
	cipherMethod string
}

type CipherMode struct {
	encryptFn func(cipher.Block, []byte) any
	decryptFn func(cipher.Block, []byte) any
}

const (
	ivLen      = aes.BlockSize
	sha2len    = 32
	cipherName = "AES"
)

var (
	availableModes = map[string]*CipherMode{
		"CBC": {
			encryptFn: func(b cipher.Block, iv []byte) any {
				return cipher.NewCBCEncrypter(b, iv)
			},
			decryptFn: func(b cipher.Block, iv []byte) any {
				return cipher.NewCBCDecrypter(b, iv)
			},
		},
		"OFB": {
			encryptFn: func(b cipher.Block, iv []byte) any {
				return cipher.NewOFB(b, iv)
			},
			decryptFn: func(b cipher.Block, iv []byte) any {
				return cipher.NewOFB(b, iv)
			},
		},
		"CFB": {
			encryptFn: func(b cipher.Block, iv []byte) any {
				return cipher.NewCFBEncrypter(b, iv)
			},
			decryptFn: func(b cipher.Block, iv []byte) any {
				return cipher.NewCFBDecrypter(b, iv)
			},
		},
		"CTR": {
			encryptFn: func(b cipher.Block, iv []byte) any {
				return cipher.NewCTR(b, iv)
			},
			decryptFn: func(b cipher.Block, iv []byte) any {
				return cipher.NewCTR(b, iv)
			},
		},
	}
	cipherLengths = []string{"128", "192", "256"}
)

func NewEncryptor(key, salt, cipherMethod string) *Encryptor {
	return &Encryptor{
		key:          key,
		salt:         salt,
		cipherMethod: cipherMethod,
	}
}

func (e *Encryptor) Encrypt(plaintext []byte, key string, salt string) ([]byte, error) {
	if plaintext == nil || len(plaintext) == 0 {
		return nil, fmt.Errorf("empty string")
	}

	pk, err := e.buildKey(key, salt)
	if err != nil {
		return nil, err
	}

	cipherMode, keySize, err := e.getCipher()
	if err != nil {
		return nil, err
	}

	iv, err := GenerateRandomBytes(aes.BlockSize)
	if err != nil {
		return nil, err
	}

	ciphertext, err := cipherMode.Encrypt(pk[:keySize], iv, Pkcs7Padding(plaintext, aes.BlockSize))
	if err != nil {
		return nil, err
	}

	hash := e.computeHmac(pk, ciphertext)
	computedData := append(iv, hash...)
	computedData = append(computedData, ciphertext...)
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(computedData)))
	base64.StdEncoding.Encode(buf, computedData)

	return buf, nil
}

func (e *Encryptor) Decrypt(encryptedText []byte, key string, salt string) ([]byte, error) {
	if encryptedText == nil || len(encryptedText) == 0 {
		return nil, fmt.Errorf("empty string")
	}

	pk, err := e.buildKey(key, salt)
	if err != nil {
		return nil, err
	}

	dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(encryptedText)))
	n, err := base64.StdEncoding.Decode(dbuf, encryptedText)
	if err != nil {
		return nil, err
	}

	c := dbuf[:n]
	if len(c) < (ivLen + sha2len) {
		return nil, fmt.Errorf("bad decrypted lenght")
	}

	ciphertext := c[ivLen+sha2len:]
	iv := c[:ivLen]

	cipherMode, keySize, err := e.getCipher()
	if err != nil {
		return nil, err
	}

	plaintext, err := cipherMode.Decrypt(pk[:keySize], iv, ciphertext)
	if err != nil {
		return nil, err
	}

	calculatedMac := e.computeHmac(pk, ciphertext)
	receivedMac := c[ivLen : ivLen+sha2len]

	if !hmac.Equal(calculatedMac, receivedMac) {
		return nil, fmt.Errorf("HMAC verification failed")
	}

	result, err := Pkcs7UnPadding(plaintext)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// parse and validate cipher from config
// returns cipher mode and hash key size
func (e *Encryptor) getCipher() (*CipherMode, int, error) {
	if e.cipherMethod == "" {
		return nil, 0, fmt.Errorf("undefined encryption cypher method")
	}

	formattedError := fmt.Errorf(
		"cypher method format is: {%v}-{%v}-{%v}",
		cipherName,
		strings.Join(cipherLengths, ","),
		strings.Join(MapKeys(availableModes), ","),
	)

	params := strings.Split(strings.ToUpper(e.cipherMethod), "-") // AES-{size}-{mode}
	if len(params) != 3 {
		return nil, 0, formattedError
	}

	if params[0] != cipherName {
		return nil, 0, formattedError
	}

	// get mode
	mode, ok := availableModes[params[2]]
	if !ok {
		return nil, 0, formattedError
	}

	for _, value := range cipherLengths {
		if value == params[1] {
			if parsed, err := strconv.Atoi(value); err == nil {
				return mode, parsed / 8, nil
			}
		}
	}

	return nil, 0, formattedError
}

func (e *Encryptor) buildKey(key string, salt string) ([]byte, error) {
	if key == "" {
		if e.key == "" {
			return nil, fmt.Errorf("undefined encryption key")
		}

		key = e.key
	}

	if salt == "" {
		if e.salt == "" {
			return nil, fmt.Errorf("undefined encryption salt")
		}

		salt = e.salt
	}

	hash := sha256.New()
	hash.Write([]byte(salt))
	hash.Write([]byte(key))
	hash.Write([]byte(salt))

	return hash.Sum(nil), nil
}

func (e *Encryptor) computeHmac(key []byte, data ...[]byte) []byte {
	h := hmac.New(sha256.New, key)
	for _, value := range data {
		h.Write(value)
	}
	return h.Sum(nil)
}

func (m *CipherMode) Encrypt(key []byte, iv []byte, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertextRaw := make([]byte, len(plaintext))
	mode := m.encryptFn(block, iv)

	switch mode := mode.(type) {
	case cipher.BlockMode:
		mode.CryptBlocks(ciphertextRaw, plaintext)
	case cipher.Stream:
		mode.XORKeyStream(ciphertextRaw, plaintext)
	}

	return ciphertextRaw, nil
}

func (m *CipherMode) Decrypt(key []byte, iv []byte, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	mode := m.decryptFn(block, iv)
	plaintext := make([]byte, len(ciphertext))

	switch mode := mode.(type) {
	case cipher.BlockMode:
		mode.CryptBlocks(plaintext, ciphertext)
	case cipher.Stream:
		mode.XORKeyStream(plaintext, ciphertext)
	}

	return plaintext, err
}
