package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func Encrypt(data, key []byte) ([]byte, error) {
	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return aesGCM.Seal(nonce, nonce, data, nil), nil
}
