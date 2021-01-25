package main

import (
	"fmt"
	"github.com/GreenStage/kingfish/internal/db/postgres"
	"github.com/GreenStage/kingfish/internal/handlers"
	"github.com/GreenStage/kingfish/pkg/aes"
	"net/http"
	"os"
	"time"
)

func main() {
	sessionEncryptionKey, err := aes.GenerateKey(32)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	router := handlers.NewRouter(handlers.Config{
		Drivers: map[string]handlers.Driver{
			"postgresql": &postgres.Driver{},
		},
		TemplatesDir:         readEnvString("TEMPLATES_DIR", "./template"),
		AssetsDir:            readEnvString("ASSETS_DIR", "./assets"),
		SessionEncryptionKey: sessionEncryptionKey,
		MaxSessionLifetime:   readEnvDuration("MAX_SESSION_LIFETIME", 5*time.Minute),
	})

	fmt.Println("server listening")
	if err := http.ListenAndServe(":8080", router); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("server terminated")
}

func readEnvString(key, defaultVal string) string {
	v := os.Getenv(key)
	if v == "" {
		v = defaultVal
	}
	return v
}

func readEnvDuration(key string, defaultVal time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	duration, err := time.ParseDuration(v)
	if err != nil {
		return defaultVal
	}
	return duration
}
