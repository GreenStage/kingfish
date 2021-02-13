package integrationtests

import (
	"database/sql"
	"fmt"
	"github.com/ory/dockertest/v3"
	"os"
)

var postgresTestMigrations = []string{
	`CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL
	)`,
	`CREATE TABLE products (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		price decimal NOT NULL
	)`,
	`CREATE TABLE orders (
		id SERIAL PRIMARY KEY,
		user_id integer NOT NULL,
		product_id integer NOT NULL,
		amount integer NOT NULL,
		CONSTRAINT fk_user
			  FOREIGN KEY(user_id) 
			  REFERENCES users(id),
		CONSTRAINT fk_product
			  FOREIGN KEY(product_id) 
			  REFERENCES products(id)
	)`,
	`INSERT INTO users (name,email) VALUES 
		('user1','usermail1@github.com'),
		('user2','usermail2@github.com'),
		('user3','usermail3@github.com')
	`,
	`INSERT INTO products (name, price) VALUES 
		('apple', '10.0'),
		('banana', '12.5'),
		('orange', '3.14159265358979323846')
	`,
	`INSERT INTO orders (user_id, product_id, amount) VALUES 
		('1','2','5'),
		('2','3','1'),
		('3','3','3'),
		('3','1','2')
	`,
}

func loadPosgresTestDB(pool *dockertest.Pool, user, pw, dbname string) (url string, cleanupFn func(), err error) {
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "postgres",
		Tag:          "13.1",
		ExposedPorts: []string{"5432"},
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + pw,
			"POSTGRES_DB=" + dbname,
		},
	})

	if err != nil {
		return "", func() {}, fmt.Errorf("could not start container: %s", err)
	}

	cleanup := func() {
		if err := pool.Purge(resource); err != nil {
			fmt.Fprintf(os.Stderr, "could not remove container: %v\n", err)
		}
	}

	postgresURL := resource.GetHostPort("5432/tcp")

	var db *sql.DB
	if err := pool.Retry(func() error {
		var err error
		fmt.Fprint(os.Stderr, "waiting for containers readiness...\n")

		dbURL := fmt.Sprintf(
			"postgres://%s:%s@%s/%s?sslmode=disable",
			user,
			pw,
			postgresURL,
			dbname,
		)

		db, err = sql.Open("postgres", dbURL)
		if err != nil {
			return err
		}

		return db.Ping()
	}); err != nil {
		return postgresURL, cleanup, fmt.Errorf("could not connect to docker: %v", err)
	}
	defer db.Close()

	for i, migration := range postgresTestMigrations {
		if _, err := db.Exec(migration); err != nil {
			return postgresURL, cleanup, fmt.Errorf("could not run migration %d: %v", i, err)
		}
	}

	return postgresURL, cleanup, nil
}
