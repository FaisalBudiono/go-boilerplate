package http

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/app"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func Middleware(e *echo.Echo) {
	e.Use(otelecho.Middleware(app.ENV().AppName))
}
