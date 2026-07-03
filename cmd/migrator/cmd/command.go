package cmd

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"FaisalBudiono/go-boilerplate/internal/app/adapter/configuration/db"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/configuration/db/seeder"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/hash"
)

type Cmd string

const (
	CmdCreate  Cmd = "create"
	CmdDown    Cmd = "down"
	CmdSeed    Cmd = "seed"
	CmdStatus  Cmd = "status"
	CmdUp      Cmd = "up"
	CmdVersion Cmd = "version"
)

type command struct {
	migrator *db.Migrator

	db *sql.DB

	defaultAdminName     string
	defaultAdminEmail    string
	defaultAdminPassword string
}

func New(
	m *db.Migrator,
	db *sql.DB,
	defaultAdminName string,
	defaultAdminEmail string,
	defaultAdminPassword string,
) *command {
	return &command{
		migrator:             m,
		db:                   db,
		defaultAdminName:     defaultAdminName,
		defaultAdminEmail:    defaultAdminEmail,
		defaultAdminPassword: defaultAdminPassword,
	}
}

func (c *command) Create(_ context.Context) error {
	fmt.Print("Type migration file name: ")

	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	input := sc.Text()

	firstArg, _, _ := strings.Cut(input, " ")

	err := c.migrator.Create(firstArg)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Migration file successfully created")

	return nil
}

func (c *command) Down(ctx context.Context) error {
	fmt.Println("Start rolling back migration...")

	err := c.migrator.Down(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Finish rolling back migration...")

	return nil
}

func (c *command) Status(ctx context.Context) error {
	return c.migrator.Status(ctx)
}

func (c *command) Up(ctx context.Context) error {
	fmt.Println("Start migrating migration...")

	err := c.migrator.Up(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Finish migrating migration...")

	return nil
}

func (c *command) Version(ctx context.Context) error {
	return c.migrator.Version(ctx)
}

func (c *command) Seed(ctx context.Context) error {
	for i, s := range c.seeders() {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			fmt.Printf("Canceled by context: %s", err)
			return err
		default:
			fmt.Printf("Syncing: %s\n", s.Name())

			err := s.Seed(ctx)
			if err != nil {
				fmt.Printf("Failed to seed #%03d: %s\n", i, err)
				return err
			}

			fmt.Println("Done")
		}
	}

	return nil
}

func (c *command) seeders() []db.Seeder {
	return []db.Seeder{
		seeder.NewAdmin(
			c.db,
			c.defaultAdminName,
			c.defaultAdminEmail,
			c.defaultAdminPassword,
			hash.NewArgon(),
		),
	}
}
