package main

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/otel"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/app"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/providers"
	"context"

	"github.com/labstack/echo/v4"
)

func main() {
	app.BindENV()

	ctx := context.Background()

	shutdown, err := otel.SetupOTelSDK(ctx)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := shutdown(ctx)
		if err != nil {
			panic(err)
		}
	}()

	tracer := otel.NewTracer(app.ENV().AppName)
	logger := otel.NewLogger(app.ENV().AppName)

	monitorings.SetUp(tracer, logger)
	providers.SetUp()

	e := echo.New()

	http.Middleware(e)
	http.Routes(e)

	e.Logger.Fatal(e.Start(":8080"))
}
