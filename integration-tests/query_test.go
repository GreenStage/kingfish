package integrationtests

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestRunQuery_401WithoutSendingToken(t *testing.T) {
	// Do a simple get tables to assert the returned token is valid
	req, err := http.NewRequest("GET", server.URL+"/tables", nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, 401, resp.StatusCode)
}

func TestRunQuery_AuthErrors(t *testing.T) {
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
			req, err := http.NewRequest("POST", server.URL+"/query", nil)
			assert.NoError(t, err)

			req.Header.Set("Authorization", test.tokenHeaderValue)
			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			assert.Equal(t, test.wantCode, resp.StatusCode)
		})
	}
}

func TestRunQuery_IdempotentRuns(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantRes  string
		wantCode int
	}{
		{
			name:     "invalid sql select",
			query:    "SELECT ** FROM",
			wantCode: 400,
			//TODO: change this, dont wanna propagate driver errors
			wantRes: `{"error":"pq: syntax error at or near \"FROM\""}`,
		},
		{
			name:  "select * from users",
			query: "SELECT * FROM users",
			wantRes: `{
				"columns": [
					{
						"name": "id",
						"type": "INT4"
					},
					{
						"name": "name",
						"type": "TEXT"
					},
					{
						"name": "email",
						"type": "TEXT"
					}
				],
				"rows":[
					[1,"user1","usermail1@github.com"],
					[2,"user2","usermail2@github.com"],
					[3,"user3","usermail3@github.com"]
				]
			}`,
		},
		{
			name: "select email from users left join products and orders",
			query: `
				SELECT u.email, p.name AS product, o.amount, o.amount * p.price AS total FROM users u
					LEFT JOIN orders o ON u.id = o.user_id
					LEFT JOIN products p ON p.id = o.product_id`,
			wantRes: `{
				"columns": [
					{
						"name": "email",
						"type": "TEXT"
					},
					{
						"name": "product",
						"type": "TEXT"
					},
					{
						"name": "amount",
						"type": "INT4"
					},
					{
						"name": "total",
						"type": "NUMERIC"
					}
				],
				"rows":[
					["usermail1@github.com","banana",5, 62.5],
					["usermail2@github.com","orange",1, "3.14159265358979323846"],
					["usermail3@github.com","orange",3, "9.42477796076937971538"],
					["usermail3@github.com","apple",2, 20.0]
				]
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token, _ := doLogin(t, getDefaultPgDBConfig())
			assert.NotEmpty(t, token)

			req, _ := http.NewRequest("POST", server.URL+"/query", strings.NewReader(test.query))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "text/plain; charset=UTF-8")

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			if test.wantCode == 0 {
				test.wantCode = 200
			}
			assert.Equal(t, test.wantCode, resp.StatusCode)

			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			strBody := string(bodyBytes)
			assert.JSONEq(t, test.wantRes, strBody)
		})
	}
}
