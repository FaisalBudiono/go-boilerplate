package db

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/env"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func PostgresConn() *sql.DB {
	return makeConnectionPostgres()
}

func makeConnectionPostgres() *sql.DB {
	source := makePostgresDSN(
		env.Get().PgUser,
		env.Get().PgPassword,
		env.Get().PgHost,
		env.Get().PgPort,
		env.Get().PgDBName,
		env.Get().PgSSLMode,
	)

	db, err := sql.Open("postgres", source)
	if err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db
}

func makePostgresDSN(
	user, password, host, port, dbName, sslMode string,
) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbName, sslMode,
	)
}
