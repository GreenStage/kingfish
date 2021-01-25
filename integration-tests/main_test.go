package integration_tests

import (
	"context"
	"fmt"
	"github.com/GreenStage/kingfish/internal/db/postgres"
	"github.com/GreenStage/kingfish/internal/handlers"
	"github.com/GreenStage/kingfish/internal/logger"
	"github.com/GreenStage/kingfish/pkg/aes"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"go.uber.org/zap"
	"net"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var (
	serverUrl string
)

type dbLoader func(pool *dockertest.Pool, user, pw, db, hostPort string) (url string, cleanupFn func(), err error)
type dbConfig struct {
	user, pass, db, url, hostport string
	loader                        dbLoader
}

var testDBs = map[string]*dbConfig{
	"postgresql": {
		user:     "usr",
		pass:     "pw",
		db:       "db",
		hostport: "8082",
		loader:   loadPosgresTestDB,
	},
}

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	pool, err := dockertest.NewPool("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not connect to docker: %s\n", err)
		return 1
	}

	for k, v := range testDBs {
		generatedUrl, cleanup, err := v.loader(pool, v.user, v.pass, v.db, v.hostport)
		if cleanup != nil {
			defer cleanup()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "could load %s db: %s\n", k, err)
			return 2
		}
		v.url = generatedUrl
	}

	sessionEncryptionKey, err := aes.GenerateKey(32)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(3)
	}

	handler := handlers.NewRouter(handlers.Config{
		Drivers: map[string]handlers.Driver{
			"postgresql": &postgres.Driver{},
		},
		TemplatesDir:         "../template",
		AssetsDir:            "../assets",
		SessionEncryptionKey: sessionEncryptionKey,
		MaxSessionLifetime:   10 * time.Second,
	})

	server := httptest.NewServer(handler)
	server.Config.ConnContext = func(ctx context.Context, c net.Conn) context.Context {
		log, _ := zap.NewDevelopment()
		return logger.ToContext(ctx, log)
	}
	serverUrl = server.URL

	return m.Run()
}