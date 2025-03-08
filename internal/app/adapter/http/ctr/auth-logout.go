package ctr

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/res"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/res/errcode"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/app/util/otel"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/trace"
)

type reqAuthLogout struct {
	ctx    context.Context
	tracer trace.Tracer

	BodyRefreshToken string `json:"refreshToken" validate:"required"`
}

func (r *reqAuthLogout) Bind(c echo.Context) error {
	msgs := make(map[string][]string, 0)

	err := c.Bind(r)
	if err != nil {
		var jsonErr *json.UnmarshalTypeError
		if !errors.As(err, &jsonErr) {
			return tracerr.Wrap(err)
		}

		if jsonErr.Field == "refreshToken" {
			msgs["refreshToken"] = append(msgs["refreshToken"], "string")
		}
	}

	err = validator.New().StructCtx(r.ctx, r)
	if err != nil {
		var valErr validator.ValidationErrors
		if !errors.As(err, &valErr) {
			return tracerr.Wrap(err)
		}

		for _, fe := range valErr {
			if fe.Field() == "BodyRefreshToken" {
				msgs["refreshToken"] = append(msgs["refreshToken"], fe.Tag())
			}
		}
	}

	if len(msgs) > 0 {
		return res.NewErrorUnprocessable(msgs)
	}

	return nil
}

func (r *reqAuthLogout) Context() context.Context {
	return r.ctx
}

func (r *reqAuthLogout) RefreshToken() string {
	return r.BodyRefreshToken
}

func AuthLogout(tracer trace.Tracer, srv *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracer.Start(c.Request().Context(), "route: logout")
		defer span.End()

		i := &reqAuthLogout{
			ctx:    ctx,
			tracer: tracer,
		}

		err := i.Bind(c)
		if err != nil {
			if unErr, ok := err.(*res.UnprocessableErrResponse); ok {
				return c.JSON(http.StatusUnprocessableEntity, unErr)
			}

			otel.SpanLogError(span, err, "binding request error")
			return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
		}

		err = srv.Logout(i)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidToken) {
				return c.JSON(
					http.StatusUnauthorized,
					res.NewError("Invalid refresh token", errcode.AuthInvalidCredentials),
				)
			}
			if errors.Is(err, auth.ErrTokenExpired) {
				return c.JSON(
					http.StatusUnauthorized,
					res.NewError("Refresh token expired", errcode.AuthInvalidCredentials),
				)
			}

			otel.SpanLogError(span, err, "error caught in service")
			return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
		}

		return c.NoContent(http.StatusNoContent)
	}
}
