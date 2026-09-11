package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"FaisalBudiono/go-boilerplate/internal/app/adapter/configuration/otel"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/in/http"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/app"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/mon"
	"FaisalBudiono/go-boilerplate/internal/app/providers"

	"github.com/labstack/echo/v5"
)

func main() {
	app.BindENV()

	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Println(err)
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

	shutdowns, err := setupStartup(ctx)
	defer func() {
		for i, sd := range shutdowns {
			err := sd.Shutdown()
			if err != nil {
				log.Printf("shutdown #%d error: %s", i, err)
			}
		}
	}()
	if err != nil {
		return err
	}

	e := echo.New()

	http.Middleware(e)
	http.Routes(e)

	err = e.Start(":8080")
	if err != nil {
		e.Logger.ErrorContext(ctx, err.Error())
		return err
	}

	return nil
}

type shutdowner interface {
	Shutdown() error
}

func setupStartup(ctx context.Context) ([]shutdowner, error) {
	ctx, span := mon.Tracer().Start(ctx, "app.main.setup-startup")
	defer span.End()

	var shutdowners []shutdowner

	sds, err := providers.Setup(ctx)
	if err != nil {
		return shutdowners, err
	}
	for _, sd := range sds {
		shutdowners = append(shutdowners, sd)
	}

	return shutdowners, nil
}
