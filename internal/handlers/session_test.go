package handlers

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"github.com/GreenStage/kingfish/internal/db"
	"github.com/GreenStage/kingfish/pkg/aes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_SessionReader_NoAuthorizationHeader_Returns401(t *testing.T) {
	router := testSessionHappyPathRouter()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "", nil)
	assert.NotPanics(t, func() {
		router.sessionReader(stubFinalHandler()).ServeHTTP(w, r)
	})

	assert.Equal(t, 401, w.Code)
}

func Test_SessionReader_NoToken_Returns401(t *testing.T) {
	router := testSessionHappyPathRouter()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "", nil)
	r.Header.Set("Authorization", "Bearer "+"")
	assert.NotPanics(t, func() {
		router.sessionReader(stubFinalHandler()).ServeHTTP(w, r)
	})

	assert.Equal(t, 401, w.Code)
}

func Test_SessionReader_InvalidB64Token_Returns401(t *testing.T) {
	router := testSessionHappyPathRouter()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "", nil)
	r.Header.Set("Authorization", "Bearer "+"invalid")
	assert.NotPanics(t, func() {
		router.sessionReader(stubFinalHandler()).ServeHTTP(w, r)
	})

	assert.Equal(t, 401, w.Code)
}

func Test_SessionReader_InvalidClientTokenLen_Returns401(t *testing.T) {
	router := testSessionHappyPathRouter()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "", nil)

	r.Header.Set("Authorization", "Bearer "+base64.StdEncoding.EncodeToString([]byte{20, 0, 0, 00, 3, 4}))
	assert.NotPanics(t, func() {
		router.sessionReader(stubFinalHandler()).ServeHTTP(w, r)
	})

	assert.Equal(t, 401, w.Code)
}

func Test_SessionReader_TokenNotEncryptedByServer_Returns401(t *testing.T) {
	router := testSessionHappyPathRouter()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "", nil)

	r.Header.Set("Authorization", "Bearer "+base64.StdEncoding.EncodeToString([]byte{2, 0, 0, 0, 3, 4, 5}))
	assert.NotPanics(t, func() {
		router.sessionReader(stubFinalHandler()).ServeHTTP(w, r)
	})

	assert.Equal(t, 401, w.Code)
}

func Test_SessionReader_SessionNotEncryptedWithClientKey_Returns401(t *testing.T) {
	router := testSessionHappyPathRouter()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "", nil)

	actualClientKey := []byte("testClienKey01234678998765434567")
	token := append([]byte{32, 0, 0, 0}, actualClientKey...)

	encryptedByServer, _ := aes.Encrypt([]byte("nonecryptedSession"), router.config.SessionEncryptionKey)
	token = append(token, encryptedByServer...)

	r.Header.Set("Authorization", "Bearer "+base64.StdEncoding.EncodeToString(token))
	assert.NotPanics(t, func() {
		router.sessionReader(stubFinalHandler()).ServeHTTP(w, r)
	})

	assert.Equal(t, 401, w.Code)
}

func Test_SessionReader_EmptyData_Returns401(t *testing.T) {
	router := testSessionHappyPathRouter()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "", nil)

	actualClientKey := []byte("testClienKey01234678998765434567")
	token := append([]byte{32, 0, 0, 0}, actualClientKey...)

	encryptedByClient, _ := aes.Encrypt([]byte(""), actualClientKey)
	encryptedByServer, _ := aes.Encrypt(encryptedByClient, router.config.SessionEncryptionKey)
	token = append(token, encryptedByServer...)

	r.Header.Set("Authorization", "Bearer "+base64.StdEncoding.EncodeToString(token))
	assert.NotPanics(t, func() {
		router.sessionReader(stubFinalHandler()).ServeHTTP(w, r)
	})

	assert.Equal(t, 401, w.Code)
}

func Test_SessionReader_AlreadyExpiredToken_Returns401(t *testing.T) {
	router := testSessionHappyPathRouter()
	token, err := router.newSessionToken(sessionData{
		IdleLifeTime: 1 * time.Second,
	})
	assert.Nil(t, err)

	time.Sleep(2 * time.Second)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "", nil)

	r.Header.Set("Authorization", "Bearer "+token)
	assert.NotPanics(t, func() {
		router.sessionReader(stubFinalHandler()).ServeHTTP(w, r)
	})

	assert.Equal(t, 401, w.Code)
}

func Test_SessionReader_InvalidSessionSchema_Returns500(t *testing.T) {
	router := testSessionHappyPathRouter()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "", nil)

	actualClientKey := []byte("testClienKey01234678998765434567")
	token := append([]byte{32, 0, 0, 0}, actualClientKey...)

	expiresAt := time.Now().Add(time.Hour).Unix()
	raw := append([]byte{0, 0, 0, 0, 0, 0, 0, 0}, []byte("[invalidstruct}")...)
	binary.LittleEndian.PutUint64(raw, uint64(expiresAt))

	encryptedByClient, _ := aes.Encrypt(raw, actualClientKey)
	encryptedByServer, _ := aes.Encrypt(encryptedByClient, router.config.SessionEncryptionKey)
	token = append(token, encryptedByServer...)

	r.Header.Set("Authorization", "Bearer "+base64.StdEncoding.EncodeToString(token))
	assert.NotPanics(t, func() {
		router.sessionReader(stubFinalHandler()).ServeHTTP(w, r)
	})

	assert.Equal(t, 500, w.Code)
}

func Test_SessionReader_UnknownDriver_Returns500(t *testing.T) {
	router := testSessionHappyPathRouter()
	token, err := router.newSessionToken(sessionData{
		IdleLifeTime: 1 * time.Minute,
		Driver:       "unknown",
	})
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "", nil)

	r.Header.Set("Authorization", "Bearer "+token)
	assert.NotPanics(t, func() {
		router.sessionReader(stubFinalHandler()).ServeHTTP(w, r)
	})

	assert.Equal(t, 500, w.Code)
}

func Test_SessionReader_ErrorConnectingToDB_Returns403(t *testing.T) {
	router := testSessionHappyPathRouter()
	driver := &MockDriver{}
	driver.On("NewConnection", db.ConnectionConfig{}).Return(nil, errors.New("boom!"))
	router.config.Drivers["testDriver"] = driver

	token, err := router.newSessionToken(sessionData{
		IdleLifeTime: 1 * time.Minute,
		Driver:       "testDriver",
	})
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "", nil)

	r.Header.Set("Authorization", "Bearer "+token)
	assert.NotPanics(t, func() {
		router.sessionReader(stubFinalHandler()).ServeHTTP(w, r)
	})

	assert.Equal(t, 403, w.Code)
}

func Test_SessionReader_SuccessInjectsSessionIntoContext(t *testing.T) {
	router := testSessionHappyPathRouter()
	conn := &MockDbConnection{}
	conn.On("Close").Return(nil)

	connConfig := db.ConnectionConfig{
		Username: "testUser123",
		Password: "testPass123",
		Dbname:   "tesDbName123",
		Host:     "testHostName123",
		Persist:  true,
	}

	driver := &MockDriver{}
	driver.On("NewConnection", connConfig).Return(conn, nil)
	router.config.Drivers["testDriver"] = driver

	token, err := router.newSessionToken(sessionData{
		Driver:       "testDriver",
		ConnConfig:   connConfig,
		IdleLifeTime: 1 * time.Minute,
	})
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "", nil)

	r.Header.Set("Authorization", "Bearer "+token)
	assert.NotPanics(t, func() {
		router.sessionReader(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := r.Context().Value(sessionCtxKey{}).(sessionData)
			assert.True(t, ok)
			w.WriteHeader(200)
		})).ServeHTTP(w, r)
	})

	assert.Equal(t, 200, w.Code)
	driver.AssertExpectations(t)
	conn.AssertExpectations(t)
}

func testSessionHappyPathRouter() router {
	conn := &MockDbConnection{}
	conn.On("Close").Return(nil)

	driver := &MockDriver{}
	driver.On("NewConnection", db.ConnectionConfig{
		Username: "testUser123",
		Password: "testPass123",
		Dbname:   "tesDbName123",
		Host:     "testHostName123",
		Persist:  true,
	}).Return(conn, nil)

	return router{
		config: Config{
			SessionEncryptionKey: []byte("12345678987654321123457789213982"),
			Drivers: map[string]Driver{
				"testDriver": driver,
			},
		},
		chiMux: nil,
	}
}

func stubFinalHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
}
