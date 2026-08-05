package http

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/mon"

	"github.com/labstack/echo/v5"
)

func requestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			_, span := mon.Tracer().Start(
				c.Request().Context(), "http.middleware.request-id-setter",
			)
			defer span.End()

			c.Response().Header().Set(
				echo.HeaderXRequestID, span.SpanContext().TraceID().String(),
			)

			return next(c)
		}
	}
}
