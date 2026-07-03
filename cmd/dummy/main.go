package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"komdigi-immigration/internal/app/adapter/configuration/db"
	"komdigi-immigration/internal/app/adapter/configuration/otel"
	"komdigi-immigration/internal/app/adapter/hardcode/passport"
	"komdigi-immigration/internal/app/core/util/app"
	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/domain"
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

	monitoring.SetUp(tracer, logger)

	ctx, span := monitoring.Tracer().Start(ctx, "dummy.main")
	defer span.End()

	adapter := passport.New()
	_ = adapter

	dbConn, err := db.PostgresConn()
	if err != nil {
		return err
	}
	_ = dbConn

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

	asd := domain.IMEI("358876627291952")

	fmt.Println("nilai")
	fmt.Println(asd)
	fmt.Printf("%t", asd.IsValid())

	fmt.Println()
	fmt.Println("done")

	return nil
}

func makau(s string) (string, error) {
	n := len(s)
	if n == 8 {
		return "", errors.New("WOI GUA BENCI 8")
	}

	return strconv.Itoa(n), nil
}

func kicauMania(isi string, err error) func() {
	return func() {
		log.Println("ISI DARI KICAU MANIA")
		log.Println(isi)
		if err != nil {
			log.Println("error KICAU")
			log.Println(err.Error())
		}
	}
}
