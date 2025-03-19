package cmd

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

const (
	migrationDir string = "db/migrations"
)

type gooseMigrator struct {
	db *sql.DB
}

func newMigrator(db *sql.DB) *gooseMigrator {
	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	return &gooseMigrator{
		db: db,
	}
}

func (g *gooseMigrator) Create(filename string) {
	err := goose.Create(g.db, migrationDir, filename, "sql")
	if err != nil {
		panic(err)
	}
}

func (g *gooseMigrator) Status() {
	err := goose.Status(g.db, migrationDir)
	if err != nil {
		panic(err)
	}
}

func (g *gooseMigrator) Version() {
	err := goose.Version(g.db, migrationDir)
	if err != nil {
		panic(err)
	}
}

func (g *gooseMigrator) Up() {
	err := goose.Up(g.db, migrationDir)
	if err != nil {
		panic(err)
	}
}

func (g *gooseMigrator) Down() {
	err := goose.Down(g.db, migrationDir)
	if err != nil {
		panic(err)
	}
}
