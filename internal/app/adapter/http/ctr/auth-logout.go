package ctr

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/res"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httputil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
)

type reqAuthLogout struct {
	tracer trace.Tracer

	ctx context.Context

	BodyRefreshToken string `json:"refreshToken" validate:"required"`
}

func (r *reqAuthLogout) Bind(c echo.Context) error {
	_, span := r.tracer.Start(r.ctx, "req: logout")
	defer span.End()

	errMsgs := make(res.VerboseMetaMsgs, 0)

	validationErr, err := httputil.Bind(r, map[string]string{
		"refreshToken": "string",
	}, c)
	if err != nil {
		otel.SpanLogError(span, err, "failed to bind")
		return err
	}
	errMsgs.AppendDomMap(validationErr)

	validationErr, err = httputil.ValidateStruct(r, map[string]string{
		"BodyRefreshToken": "refreshToken",
	})
	if err != nil {
		otel.SpanLogError(span, err, "unhandled validator error")
		return err
	}
	errMsgs.AppendDomMap(validationErr)

	if len(errMsgs) > 0 {
		return res.NewErrorUnprocessableVerbose(errMsgs)
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
