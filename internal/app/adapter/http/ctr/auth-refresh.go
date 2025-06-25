package ctr

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/res"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httputil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type reqAuthRefreshToken struct {
	ctx context.Context

	BodyRefreshToken string `json:"refreshToken" validate:"required"`
}

func (r *reqAuthRefreshToken) Bind(c echo.Context) error {
	_, span := monitorings.Tracer().Start(r.ctx, "req: refresh token")
	defer span.End()

	errMsgs := make(res.VerboseMetaMsgs, 0)

	validationErr, err := httputil.Bind(c, r, map[string]string{
		"refreshToken": "string",
	})
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
		return res.NewErrorUnprocessable(errMsgs)
	}

	return nil
}

func (r *reqAuthRefreshToken) Context() context.Context {
	return r.ctx
}

func (r *reqAuthRefreshToken) RefreshToken() string {
	return r.BodyRefreshToken
}

func AuthRefresh(srv *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := monitorings.Tracer().Start(c.Request().Context(), "route: refresh token")
		defer span.End()

		i := &reqAuthRefreshToken{
			ctx: ctx,
		}

		err := i.Bind(c)
		if err != nil {
			if unErr, ok := err.(*res.UnprocessableErrResponse); ok {
				return c.JSON(http.StatusUnprocessableEntity, unErr)
			}
			otel.SpanLogError(span, err, "binding request error")

			return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
		}

		token, err := srv.RefreshToken(i)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidToken) {
				return c.JSON(
					http.StatusUnauthorized,
					res.NewError("Invalid refresh token", errcode.AuthInvalidCredentials),
				)
			}
			otel.SpanLogError(span, err, "error caught in service")

			return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
		}

		return c.JSON(http.StatusOK, res.Auth(token))
	}
}
