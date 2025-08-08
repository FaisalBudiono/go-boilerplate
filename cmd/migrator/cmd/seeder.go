package cmd

import (
	"FaisalBudiono/go-boilerplate/cmd/migrator/seeder"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/db"
	"context"
	"fmt"
)

func DBSeed() {
	fmt.Println("Running seeder...")

	var err error
	defer func() {
		if err != nil {
			fmt.Println("Running seeder FAILED...")
			fmt.Printf("Reason:\n%s", err)
		}

		fmt.Println("Running seeder SUCCESS...")
	}()

	conn := db.PostgresConn()
	ctx := context.Background()

	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		return
	}
	defer tx.Rollback()

	seeds := []seeder.Seeder{
		seeder.NewSuperAdmin(ctx, tx),
	}

	for _, s := range seeds {
		fmt.Printf("Seeding %s START\n", s.Name())
		defer fmt.Printf("Seeding %s END\n", s.Name())

		err = s.Seed()
		if err != nil {
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		return
	}
}
