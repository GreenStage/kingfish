package integration_tests

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/cookiejar"
	"testing"
)

func getClient(t *testing.T) *http.Client {
	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)
	return &http.Client{Jar: jar}
}
