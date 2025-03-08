package cmd

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/db"
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	CmdCreate  string = "create"
	CmdDBSeed  string = "db:seed"
	CmdDown    string = "down"
	CmdStatus  string = "status"
	CmdUp      string = "up"
	CmdVersion string = "version"
)

func Create() {
	fmt.Print("Type migration file name: ")

	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	input := sc.Text()

	firstArg := strings.Split(input, " ")[0]

	newMigrator(db.PostgresConn()).Create(firstArg)

	fmt.Println()
	fmt.Println("Migration file successfully created")
}

func Down() {
	fmt.Println("Start rolling back migration...")
	newMigrator(db.PostgresConn()).Down()
	fmt.Println("Finish rolling back migration...")
}

func Status() {
	newMigrator(db.PostgresConn()).Status()
}

func Up() {
	fmt.Println("Start migrating migration...")
	newMigrator(db.PostgresConn()).Up()
	fmt.Println("Finish migrating migration...")
}

func Version() {
	newMigrator(db.PostgresConn()).Version()
}
