package authctr

import (
	"net/http"

	"komdigi-immigration/internal/app/adapter/in/http/req"
	"komdigi-immigration/internal/app/adapter/in/http/res"
	"komdigi-immigration/internal/app/core/util/httpfmt"
	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"

	"github.com/labstack/echo/v5"
)

func Userinfo() echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, span := monitoring.Tracer().Start(c.Request().Context(), "http.route.auth.userinfo")
		defer span.End()

		actor, err := req.AuthUser(ctx)
		if err != nil {
			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to get authenticated user"),
			)
			return c.JSON(http.StatusInternalServerError, httpfmt.NewErrorGeneric(err, httpfmt.WithTraceID(span)))
		}

		return c.JSON(http.StatusOK, res.User(actor.User))
	}
}
