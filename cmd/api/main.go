package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"FaisalBudiono/go-boilerplate/internal/app/adapter/configuration/otel"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/in/http"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/app"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/mon"
	"FaisalBudiono/go-boilerplate/internal/app/providers"

	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/trace/noop"
)

func main() {
	app.BindENV()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	mon.SetUp(
		noop.Tracer{},
		slog.New(slog.NewJSONHandler(io.Discard, nil)),
	)

	if err := run(ctx); err != nil {
		mon.Stdout().ErrorContext(ctx, "error caught on", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(errors.New("from main defer"))

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
				mon.Stdout().ErrorContext(
					ctx, fmt.Sprintf("shutdown #%d error", i), slog.Any("err", err),
				)
			}
		}
	}()
	if err != nil {
		return err
	}

	e := echo.New()

	http.Middleware(e)
	http.Routes(e)

	err = echo.StartConfig{Address: ":8080"}.Start(ctx, e)
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
