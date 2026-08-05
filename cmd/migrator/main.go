package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"FaisalBudiono/go-boilerplate/cmd/migrator/cmd"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/configuration/db"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/configuration/otel"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/app"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/mon"
)

func main() {
	app.BindENV()

	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Printf("error caught on: %s", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	shutdown, err := otel.SetupOTelSDK(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err := shutdown(ctx)
		if err != nil {
			fmt.Println(err)
		}
	}()

	tracer := otel.NewTracer(app.ENV().AppName)
	logger := otel.NewLogger(app.ENV().AppName)

	mon.SetUp(tracer, logger)

	ctx, span := mon.Tracer().Start(ctx, "migrator.main")
	defer span.End()

	args := os.Args
	if len(args) == 1 {
		helpScreen()
		return nil
	}

	dbConn, err := db.PostgresConn()
	if err != nil {
		return err
	}
	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Printf("failed to close db connection: %s", err)
		}
	}()

	migrator, err := db.NewMigrator(dbConn)
	if err != nil {
		return err
	}

	cmds := cmdList(dbConn, migrator)
	command := cmd.Cmd(args[1])

	runner, ok := cmds[command]
	if !ok {
		helpScreen()
		return nil
	}

	err = runner(ctx)
	if err != nil {
		log.Printf("error caught on: %s", err)
		return err
	}

	return nil
}

type cmdRunner func(ctx context.Context) error

func cmdList(
	db *sql.DB,
	migrator *db.Migrator,
) map[cmd.Cmd]cmdRunner {
	c := cmd.New(
		migrator, db,
		app.ENV().Seeder.Admin.Name,
		app.ENV().Seeder.Admin.Email,
		app.ENV().Seeder.Admin.Password,
	)

	return map[cmd.Cmd]cmdRunner{
		cmd.CmdCreate:  c.Create,
		cmd.CmdDown:    c.Down,
		cmd.CmdStatus:  c.Status,
		cmd.CmdUp:      c.Up,
		cmd.CmdVersion: c.Version,
		cmd.CmdSeed:    c.Seed,
	}
}

func helpScreen() {
	fmt.Printf("Should keyin valid command:\n")

	cmds := []cmd.Cmd{
		cmd.CmdCreate,
		cmd.CmdUp,
		cmd.CmdDown,
		cmd.CmdStatus,
		cmd.CmdVersion,
		cmd.CmdSeed,
	}

	for _, cmd := range cmds {
		fmt.Printf("    - %s\n", cmd)
	}
}
