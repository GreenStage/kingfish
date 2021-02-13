package postgres

import (
	"fmt"
	"github.com/GreenStage/kingfish/internal/db"
	"strconv"
	"strings"

	// sql driver
	_ "github.com/lib/pq"
)

const (
	defaultPort = 5432
)

// Driver Postgres driver
type Driver struct {
}

func (d *Driver) NewConnection(config db.ConnectionConfig) (db.Connection, error) {
	var err error
	hostAndPort := strings.Split(config.Host, ":")

	port := defaultPort
	if len(hostAndPort) > 1 {
		port, err = strconv.Atoi(hostAndPort[1])
		if err != nil {
			return nil, fmt.Errorf("could not parse port from hostname: %w", err)
		}
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		hostAndPort[0], port, config.Username, config.Password, config.Dbname)

	return connect(psqlInfo)
}
