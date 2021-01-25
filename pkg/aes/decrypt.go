package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

func Decrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()

	if len(data) < nonceSize {
		return nil, errors.New("invalid data len")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	return aesGCM.Open(nil, nonce, ciphertext, nil)
}
