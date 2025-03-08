package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func PostgresConn() *sql.DB {
	return makeConnectionPostgres()
}

func makeConnectionPostgres() *sql.DB {
	source := makePostgresDSN(
		viper.GetString("POSTGRES_USER"),
		viper.GetString("POSTGRES_PASSWORD"),
		viper.GetString("POSTGRES_HOST"),
		viper.GetString("POSTGRES_PORT"),
		viper.GetString("POSTGRES_DB_NAME"),
		viper.GetString("POSTGRES_SSL_MODE"),
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

