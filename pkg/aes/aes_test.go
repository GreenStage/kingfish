package aes

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_EncryptsAndDecryptsDataCorrectly(t *testing.T) {
	keyBytes := []byte("GME goes brrrrrrrrrrrrrrrrrrrrrrr")

	key, err := GenerateKeyFromReader(32, bytes.NewReader(keyBytes))
	assert.NoError(t, err)

	data := "i like the stock"

	encrypted, err := Encrypt([]byte(data), key)
	assert.NoError(t, err)
	assert.NotNil(t, encrypted)
	assert.NotEmpty(t, encrypted)

	decrypted, err := Decrypt(encrypted, key)
	assert.NoError(t, err)
	assert.Equal(t, []byte(data), decrypted)
}

func Test_GenerateKeyErrorOnKeySmallerThanSize(t *testing.T) {
	keyBytes := []byte("GME goes brrr")

	_, err := GenerateKeyFromReader(32, bytes.NewReader(keyBytes))
	assert.NotNil(t, err)
}

func Test_EncryptsReturnsErrorOnKeyLessThan32Bytes(t *testing.T) {
	keyBytes := []byte("GME goes brrrrrrrrrrrrrrrrrrrrrrrrrr")

	key, err := GenerateKeyFromReader(31, bytes.NewReader(keyBytes))
	assert.NoError(t, err)

	data := "i like the stock"

	_, err = Encrypt([]byte(data), key)
	assert.NotNil(t, err)
}

func Test_DecryptsReturnsInvalidKeyOnSmallKey(t *testing.T) {
	keyBytes := []byte("GME goes brrrrrrrrrrrrrrrrrrrrrrrrrr")

	key, err := GenerateKeyFromReader(31, bytes.NewReader(keyBytes))
	assert.NoError(t, err)

	data := "i like the stock"

	_, err = Decrypt([]byte(data), key)
	assert.NotNil(t, err)
}
