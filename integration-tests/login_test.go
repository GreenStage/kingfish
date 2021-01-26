package integration_tests

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/url"
	"strings"
	"testing"
)

func TestGetLoginPage_ContainsRequiredFields(t *testing.T) {
	r, err := getClient(t).Get(serverUrl + "/login")
	assert.NoError(t, err)

	defer r.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	assert.Contains(t, string(bodyBytes), "name=\"driver\"")
	assert.Contains(t, string(bodyBytes), "name=\"hostname\"")
	assert.Contains(t, string(bodyBytes), "name=\"username\"")
	assert.Contains(t, string(bodyBytes), "name=\"password\"")
	assert.Contains(t, string(bodyBytes), "name=\"db\"")
}

func TestPostLoginWithPostgresDriver(t *testing.T) {
	conf := testDBs["postgresql"]
	tests := []struct {
		name        string
		contentType string
		form        url.Values
		wantCode    int
		wantInBody  string
	}{
		{
			name: "successful login returns home page",
			form: url.Values{
				"hostname": []string{conf.url},
				"username": []string{conf.user},
				"password": []string{conf.pass},
				"db":       []string{conf.db},
			},
			contentType: "application/x-www-form-urlencoded",
			wantCode:    200,
			wantInBody:  "Hello " + conf.user,
		},
		{
			name: "wrong hostname returns 400",
			form: url.Values{
				"hostname": []string{"invalid:5432"},
				"username": []string{conf.user},
				"password": []string{conf.pass},
				"db":       []string{conf.db},
			},
			contentType: "application/x-www-form-urlencoded",
			wantCode:    400,
		},
		{
			name: "wrong username returns 400",
			form: url.Values{
				"hostname": []string{conf.url},
				"username": []string{"invalid"},
				"password": []string{conf.pass},
				"db":       []string{conf.db},
			},
			contentType: "application/x-www-form-urlencoded",
			wantCode:    400,
		},
		{
			name: "wrong password returns 400",
			form: url.Values{
				"hostname": []string{conf.url},
				"username": []string{conf.user},
				"password": []string{"invalid"},
				"db":       []string{conf.db},
			},
			contentType: "application/x-www-form-urlencoded",
			wantCode:    400,
		},
		{
			name: "wrong db name returns 400",
			form: url.Values{
				"hostname": []string{conf.url},
				"username": []string{conf.user},
				"password": []string{conf.pass},
				"db":       []string{"invalid"},
			},
			contentType: "application/x-www-form-urlencoded",
			wantCode:    400,
		},
		{
			name: "invalid content type returns 415",
			form: url.Values{
				"hostname": []string{conf.url},
				"username": []string{conf.user},
				"password": []string{conf.pass},
				"db":       []string{conf.db},
			},
			contentType: "invalid",
			wantCode:    415,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.form["driver"] = []string{"postgresql"}

			r, err := getClient(t).Post(serverUrl+"/login", test.contentType, strings.NewReader(test.form.Encode()))
			assert.NoError(t, err)

			defer r.Body.Close()
			bodyBytes, _ := ioutil.ReadAll(r.Body)

			assert.Equal(t, test.wantCode, r.StatusCode)
			if test.wantInBody != "" {
				assert.Contains(t, string(bodyBytes), test.wantInBody)
			}
		})
	}
}
