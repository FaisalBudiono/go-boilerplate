package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"komdigi-immigration/internal/app/adapter/configuration/otel"
	"komdigi-immigration/internal/app/adapter/in/http"
	"komdigi-immigration/internal/app/core/util/app"
	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/providers"

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

	monitoring.SetUp(tracer, logger)

	ctx, span := monitoring.Tracer().Start(ctx, "app.main")
	defer span.End()

	shutdowns, err := providers.Setup(ctx)
	if err != nil {
		return err
	}
	defer func() {
		for i, sd := range shutdowns {
			err := sd()
			if err != nil {
				otelutil.SpanLogError(
					span, err,
					otelutil.WithErrorLog(ctx),
					otelutil.WithMessage(fmt.Sprintf("failed to shutdown provider #%d", i)),
				)
			}
		}
	}()

	e := echo.New()

	http.Middleware(e)
	http.Routes(e)

	err = e.Start(":8080")
	e.Logger.ErrorContext(ctx, err.Error())

	return err
}
