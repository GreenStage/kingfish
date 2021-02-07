package integration_tests

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetTables_401WithoutSendingToken(t *testing.T) {
	// Do a simple get tables to assert the returned token is valid
	req, err := http.NewRequest("GET", server.URL+"/tables", nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, 401, resp.StatusCode)
}

func TestGetTables_AuthErrors(t *testing.T) {
	tests := []struct {
		name             string
		tokenHeaderValue string
		wantCode         int
	}{
		{
			name:             "empty header returns 401",
			tokenHeaderValue: "",
			wantCode:         401,
		},
		{
			name:             "invalid authorization type returns 401",
			tokenHeaderValue: "Basic xyz",
			wantCode:         401,
		},
		{
			name:             "empty token returns 401",
			tokenHeaderValue: "Bearer",
			wantCode:         401,
		},
		{
			name:             "invalid token returns 401",
			tokenHeaderValue: "Bearer ~~saçssasx",
			wantCode:         401,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", server.URL+"/tables", nil)
			assert.NoError(t, err)

			req.Header.Set("Authorization", test.tokenHeaderValue)
			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			assert.Equal(t, test.wantCode, resp.StatusCode)
		})
	}
}

func TestGetTablesWithPostgresDriver_SuccessReturnsTablesInfo(t *testing.T) {
	conf := testDBs["postgresql"]

	token, _ := doLogin(t, map[string]interface{}{
		"driver":   "postgresql",
		"hostname": conf.url,
		"username": conf.user,
		"password": conf.pass,
		"dbname":   conf.db,
	})
	assert.NotEmpty(t, token)

	// Do a simple get tables to assert the returned token is valid
	req, err := http.NewRequest("GET", server.URL+"/tables", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	strBody := string(bodyBytes)

	assert.Contains(t, strBody, "\"name\":\"products\"")
	assert.Contains(t, strBody, "\"name\":\"orders\"")
	assert.Contains(t, strBody, "\"name\":\"users\"")
}
