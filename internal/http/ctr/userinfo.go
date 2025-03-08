package ctr

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/http/req"
	"FaisalBudiono/go-boilerplate/internal/http/res"
	"FaisalBudiono/go-boilerplate/internal/http/res/errcode"
	"FaisalBudiono/go-boilerplate/internal/otel"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
)

func Userinfo(tracer trace.Tracer, srv *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracer.Start(c.Request().Context(), "route: userinfo")
		defer span.End()

		u, err := req.ParseToken(ctx, c, srv)
		if err != nil {
			isTokenNotProvidedErr := errors.Is(err, req.ErrInvalidToken) ||
				errors.Is(err, req.ErrNoTokenProvided) ||
				errors.Is(err, req.ErrTokenExpired)

			if isTokenNotProvidedErr {
				return c.JSON(http.StatusUnauthorized, res.NewError(err.Error(), errcode.AuthUnauthorized))
			}
			otel.SpanLogError(span, err, "error when parsing token")

			return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
		}

		return c.JSON(http.StatusOK, res.ToUser(u))
	}
}
