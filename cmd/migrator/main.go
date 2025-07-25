package main

import (
	"FaisalBudiono/go-boilerplate/cmd/migrator/cmd"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/app"
	"fmt"
	"os"
)

func main() {
	app.BindENV()

	args := os.Args
	if len(args) == 1 {
		helpScreen()
		os.Exit(0)
	}

	switch args[1] {
	case cmd.CmdCreate:
		cmd.Create()
	case cmd.CmdDBSeed:
		cmd.DBSeed()
	case cmd.CmdDown:
		cmd.Down()
	case cmd.CmdStatus:
		cmd.Status()
	case cmd.CmdUp:
		cmd.Up()
	case cmd.CmdVersion:
		cmd.Version()
	default:
		helpScreen()
	}
}

func helpScreen() {
	fmt.Printf("Should keyin valid command:\n")
	fmt.Printf("    - %s\n", cmd.CmdCreate)
	fmt.Printf("    - %s\n", cmd.CmdDBSeed)
	fmt.Printf("    - %s\n", cmd.CmdDown)
	fmt.Printf("    - %s\n", cmd.CmdStatus)
	fmt.Printf("    - %s\n", cmd.CmdUp)
	fmt.Printf("    - %s\n", cmd.CmdVersion)
}
