package aes

import (
	"crypto/rand"
	"io"
)

func GenerateKey(size int) ([]byte, error) {
	clientKey := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, clientKey)
	return clientKey, err
}
