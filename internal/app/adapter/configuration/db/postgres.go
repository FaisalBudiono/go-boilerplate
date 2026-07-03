package db

import (
	"database/sql"
	"fmt"
	"time"

	"komdigi-immigration/internal/app/core/util/app"

	"github.com/XSAM/otelsql"
	_ "github.com/lib/pq"
)

func PostgresConn() (*sql.DB, error) {
	return makeConnectionPostgres()
}

func makeConnectionPostgres() (*sql.DB, error) {
	source := makePostgresDSN(
		app.ENV().DB.Postgres.User,
		app.ENV().DB.Postgres.Password,
		app.ENV().DB.Postgres.Host,
		app.ENV().DB.Postgres.Port,
		app.ENV().DB.Postgres.DBName,
		app.ENV().DB.Postgres.SSLMode,
	)

	db, err := otelsql.Open(
		"postgres", source,
		otelsql.WithAttributes(otelsql.AttributesFromDSN(source)...),
	)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}

func makePostgresDSN(
	user, password, host, port, dbName, sslMode string,
) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbName, sslMode,
	)
}
