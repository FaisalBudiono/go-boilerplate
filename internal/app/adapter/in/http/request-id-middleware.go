package http

import (
	"komdigi-immigration/internal/app/core/util/monitoring"

	"github.com/labstack/echo/v5"
)

func requestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			_, span := monitoring.Tracer().Start(
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
