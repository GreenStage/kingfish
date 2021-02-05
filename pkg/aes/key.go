package aes

import (
	"crypto/rand"
	"io"
)

func GenerateKey(size int) ([]byte, error) {
	return GenerateKeyFromReader(size, rand.Reader)
}

func GenerateKeyFromReader(size int, reader io.Reader) ([]byte, error) {
	clientKey := make([]byte, size)
	_, err := io.ReadFull(reader, clientKey)
	return clientKey, err
}
