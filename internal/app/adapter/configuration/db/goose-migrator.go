package db

import (
	"context"
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

const (
	migrationDirCreate string = "internal/app/adapter/configuration/db/migrations"
	migrationDirEmbed  string = "migrations"
)

//go:embed all:migrations
var migrationEmbeddedBase embed.FS

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) (*Migrator, error) {
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}

	return &Migrator{
		db: db,
	}, nil
}

func (g *Migrator) Create(filename string) error {
	return goose.Create(g.db, migrationDirCreate, filename, "sql")
}

func (g *Migrator) Status(ctx context.Context) error {
	goose.SetBaseFS(migrationEmbeddedBase)

	return goose.StatusContext(ctx, g.db, migrationDirEmbed)
}

func (g *Migrator) Version(ctx context.Context) error {
	goose.SetBaseFS(migrationEmbeddedBase)

	return goose.VersionContext(ctx, g.db, migrationDirEmbed)
}

func (g *Migrator) Up(ctx context.Context) error {
	goose.SetBaseFS(migrationEmbeddedBase)

	return goose.UpContext(ctx, g.db, migrationDirEmbed)
}

func (g *Migrator) Down(ctx context.Context) error {
	goose.SetBaseFS(migrationEmbeddedBase)

	return goose.DownContext(ctx, g.db, migrationDirEmbed)
}
