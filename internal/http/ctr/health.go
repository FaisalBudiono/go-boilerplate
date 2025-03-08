package ctr

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/ht"
	"FaisalBudiono/go-boilerplate/internal/otel"
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
)

type healthRes struct {
	Ok  bool   `json:"ok"`
	Msg string `json:"message,omitempty"`
}

type healthReq struct {
	ctx context.Context
}

func (r *healthReq) Context() context.Context {
	return r.ctx
}

func Health(tracer trace.Tracer, l *slog.Logger, srv *ht.Healthcheck) echo.HandlerFunc {
	return func(c echo.Context) error {
        ctx, span := tracer.Start(c.Request().Context(), "route: healthcheck")
		defer span.End()

		err := srv.Healthcheck(&healthReq{
			ctx: ctx,
		})
		if err != nil {
			otel.SpanLogError(span, err, "healthcheck error")

			return c.JSON(http.StatusInternalServerError, healthRes{
				Ok:  false,
				Msg: err.Error(),
			})
		}

		return c.JSON(http.StatusOK, healthRes{
			Ok: true,
		})
	}
}
