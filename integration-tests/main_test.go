package integration_tests

import (
	"context"
	"fmt"
	"github.com/GreenStage/kingfish/internal/db/postgres"
	"github.com/GreenStage/kingfish/internal/handlers"
	"github.com/GreenStage/kingfish/pkg/aes"
	"github.com/GreenStage/kingfish/pkg/logger"
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
	server *httptest.Server
)

type dbLoader func(pool *dockertest.Pool, user, pw, db string) (url string, cleanupFn func(), err error)
type dbConfig struct {
	user, pass, db, url string
	loader              dbLoader
}

var testDBs = map[string]*dbConfig{
	"postgresql": {
		user:   "usr",
		pass:   "pw",
		db:     "db",
		loader: loadPosgresTestDB,
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
		generatedUrl, cleanup, err := v.loader(pool, v.user, v.pass, v.db)
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
		return 3
	}

	handler := handlers.NewRouter(handlers.Config{
		Drivers: map[string]handlers.Driver{
			"postgresql": &postgres.Driver{},
		},
		SessionEncryptionKey: sessionEncryptionKey,
		SessionIdleLifetime:  10 * time.Second,
	})

	server = httptest.NewUnstartedServer(handler)
	server.Config.ConnContext = func(ctx context.Context, c net.Conn) context.Context {
		log, _ := zap.NewDevelopment()
		return logger.ToContext(ctx, log)
	}
	server.Start()

	return m.Run()
}
