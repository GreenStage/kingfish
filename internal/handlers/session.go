package handlers

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/GreenStage/kingfish/internal/db"
	"github.com/GreenStage/kingfish/pkg/aes"
	"github.com/GreenStage/kingfish/pkg/logger"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	clientKeyLen              = 32
	reqAuthorizationHeaderKey = "Authorization"
	respTokenHeaderKey        = "X-Access-Token"
	respTokenExpInHeaderKey   = "X-Access-Token-Expires-In"
)

type sessionData struct {
	Driver       string
	ConnConfig   db.ConnectionConfig
	Conn         db.Connection
	IdleLifeTime time.Duration
}

func (r *router) newSessionToken(data sessionData) (string, error) {
	dataBytes := make([]byte, 8)

	expiresAt := time.Now().Add(data.IdleLifeTime)
	binary.LittleEndian.PutUint64(dataBytes, uint64(expiresAt.Unix()))

	confBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("could not marshal session config into valid json: %v", err)
	}
	dataBytes = append(dataBytes, confBytes...)

	clientKey, err := aes.GenerateKey(clientKeyLen)
	if err != nil {
		return "", fmt.Errorf("could not generate client key: %v", err)
	}

	// encrypt first with the generated client key
	clientEncryptedSess, err := aes.Encrypt(dataBytes, clientKey)
	if err != nil {
		return "", fmt.Errorf("could not encrypt session config with server key: %v", err)
	}

	// encrypt with the server key
	serverEncryptedSess, err := aes.Encrypt(clientEncryptedSess, r.config.SessionEncryptionKey)
	if err != nil {
		return "", fmt.Errorf("could not encrypt session config with server key: %v", err)
	}

	tokenBytes := make([]byte, 4, clientKeyLen+len(serverEncryptedSess)+4)
	binary.LittleEndian.PutUint32(tokenBytes, clientKeyLen)

	tokenBytes = append(tokenBytes, clientKey...)
	tokenBytes = append(tokenBytes, serverEncryptedSess...)

	base64Token := make([]byte, base64.StdEncoding.EncodedLen(len(tokenBytes)))
	base64.StdEncoding.Encode(base64Token, tokenBytes)

	return string(base64Token), nil
}

func (r *router) writeSessionData(w http.ResponseWriter, data sessionData) error {
	token, err := r.newSessionToken(data)
	if err != nil {
		return err
	}

	w.Header().Set(respTokenHeaderKey, token)
	w.Header().Set(respTokenExpInHeaderKey, strconv.Itoa(int(data.IdleLifeTime.Seconds())))
	return nil
}

func (r *router) sessionReader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		authHeader := strings.Split(req.Header.Get(reqAuthorizationHeaderKey), " ")
		if len(authHeader) < 2 || strings.ToLower(authHeader[0]) != "bearer" {
			logger.FromContext(ctx).Debug("invalid authorization header")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		log := logger.FromContext(ctx).With(zap.String(reqAuthorizationHeaderKey, authHeader[1]))

		rawToken, err := base64.StdEncoding.DecodeString(authHeader[1])
		if err != nil {
			log.Debug("could not decode base64 token"+authHeader[1], zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if len(rawToken) < 4 {
			log.Debug("invalid token len: " + strconv.Itoa(len(rawToken)))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		clientKeyLen := int(binary.LittleEndian.Uint32(rawToken[:4]))
		if clientKeyLen >= len(rawToken)-4 {
			log.Debug("invalid client key length in token: " + strconv.Itoa(clientKeyLen))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		clientKey := rawToken[4 : 4+clientKeyLen]
		serverEncryptedSession := rawToken[4+clientKeyLen:]

		clientEncryptedSession, err := aes.Decrypt(serverEncryptedSession, r.config.SessionEncryptionKey)
		if err != nil {
			log.Debug("could not decrypt session config with server key", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		rawSession, err := aes.Decrypt(clientEncryptedSession, clientKey)
		if err != nil {
			log.Debug("could not decrypt session config with client key", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if rawSessionLen := len(rawSession); rawSessionLen < 8 {
			log.Debug("invalid decrypted session buffer with len " + strconv.Itoa(rawSessionLen))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if expiresAt := time.Unix(int64(binary.LittleEndian.Uint64(rawSession[:8])), 0); time.Now().After(expiresAt) {
			log.Debug("token already expired")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var data sessionData
		if err := json.Unmarshal(rawSession[8:], &data); err != nil {
			log.Debug("could not unmarshal session config from json", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		driver, ok := r.config.Drivers[data.Driver]
		if !ok {
			log.Debug("unknown driver " + data.Driver)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		conn, err := driver.NewConnection(data.ConnConfig)
		if err != nil {
			log.Debug("could not establish db connection", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		defer func() {
			if e := conn.Close(); e != nil {
				log.Debug("could not close db connection", zap.Error(e))
			}
		}()

		if err := r.writeSessionData(w, data); err != nil {
			log.Debug("could not write a fresh session data", zap.Error(err))
		}

		data.Conn = conn

		req = req.WithContext(context.WithValue(ctx, sessionCtxKey, data))
		next.ServeHTTP(w, req)
	})
}
