package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"FaisalBudiono/go-boilerplate/internal/app/adapter/configuration/otel"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/app"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/mon"
)

func main() {
	app.BindENV()

	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Println("failed to run dummy", err)
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

	ctx, span := mon.Tracer().Start(ctx, "dummy.main")
	defer span.End()

	// icl := domain.NewImmigrationClearanceLogData(
	// 	[]byte(`{"status":"OK","data":{"id":"1234567890"}}`),
	// 	200,
	// 	new("ucul"),
	// 	new("dumm"),
	// 	new(domain.GenderMale),
	// 	new("bogor"),
	// 	new(time.Now().AddDate(-29, -7, -5)),
	// 	new("1234567890"),
	// 	new(domain.CountryCode("IDN")),
	// 	new(true),
	// 	new(time.Now().AddDate(-1, -2, -3)),
	// 	new(time.Now().AddDate(4, -2, -3)),
	// 	new("5"),
	// 	"1234567890",
	// )

	// icl := domain.NewImmigrationClearanceLog(
	// 	[]byte(`{"status":"OK","data":{"id":"1234567890"}}`),
	// 	500,
	// 	nil,
	// 	nil,
	// 	nil,
	// 	nil,
	// 	nil,
	// 	nil,
	// 	nil,
	// 	nil,
	// 	nil,
	// 	nil,
	// 	new("3"),
	// 	"1234567890",
	// )

	return nil
}
