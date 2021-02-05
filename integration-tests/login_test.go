package integration_tests

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestPostLoginWithPostgresDriver_InvalidJsonReturns400(t *testing.T) {
	r, err := http.Post(serverUrl+"/login", "application/json", strings.NewReader("{{}"))
	assert.NoError(t, err)

	_, err = io.Copy(ioutil.Discard, r.Body)
	assert.NoError(t, err)
	assert.NoError(t, r.Body.Close())

	assert.Equal(t, 400, r.StatusCode)
}

func TestPostLoginWithPostgresDriver_InvalidCredentials(t *testing.T) {
	conf := testDBs["postgresql"]
	tests := []struct {
		name     string
		reqBody  interface{}
		wantCode int
	}{
		{
			name: "wrong hostname returns 400",
			reqBody: map[string]interface{}{
				"driver":   "postgresql",
				"hostname": "invalid:5432",
				"username": conf.user,
				"password": conf.pass,
				"dbname":   conf.db,
			},
			wantCode: 400,
		},
		{
			name: "wrong username returns 400",
			reqBody: map[string]interface{}{
				"driver":   "postgresql",
				"hostname": conf.url,
				"username": "invalid",
				"password": conf.pass,
				"dbname":   conf.db,
			},
			wantCode: 400,
		},
		{
			name: "wrong password returns 400",
			reqBody: map[string]interface{}{
				"driver":   "postgresql",
				"hostname": conf.url,
				"username": conf.user,
				"password": "password",
				"dbname":   conf.db,
			},
			wantCode: 400,
		},
		{
			name: "wrong db name returns 400",
			reqBody: map[string]interface{}{
				"driver":   "postgresql",
				"hostname": conf.url,
				"username": conf.user,
				"password": conf.pass,
				"dbname":   "invalid",
			},
			wantCode: 400,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			body, err := json.Marshal(test.reqBody)
			assert.Nil(t, err)

			r, err := http.Post(serverUrl+"/login", "application/json", bytes.NewReader(body))
			assert.NoError(t, err)

			_, err = io.Copy(ioutil.Discard, r.Body)
			assert.NoError(t, err)
			assert.NoError(t, r.Body.Close())

			assert.Equal(t, test.wantCode, r.StatusCode)
		})
	}
}

func TestPostLoginWithPostgresDriver_SuccessReturnsValidTokenAndExpiryDate(t *testing.T) {
	conf := testDBs["postgresql"]
	token, expiresIn := doLogin(t, map[string]interface{}{
		"driver":   "postgresql",
		"hostname": conf.url,
		"username": conf.user,
		"password": conf.pass,
		"dbname":   conf.db,
	})

	assert.NotEmpty(t, token)
	assert.Greater(t, int(expiresIn), 0)

	// Do a simple get tables to assert the returned token is valid
	req, err := http.NewRequest("GET", serverUrl+"/tables", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)
}

func doLogin(t *testing.T, input interface{}) (string, float64) {
	body, err := json.Marshal(input)
	assert.NoError(t, err)

	r, err := http.Post(serverUrl+"/login", "application/json", bytes.NewReader(body))
	assert.NoError(t, err)

	data := make(map[string]interface{})
	err = json.NewDecoder(r.Body).Decode(&data)

	r.Body.Close()
	assert.NoError(t, err)

	token, _ := data["access_token"].(string)

	expiresIn, _ := data["expires_in"].(float64)

	return token, expiresIn
}
