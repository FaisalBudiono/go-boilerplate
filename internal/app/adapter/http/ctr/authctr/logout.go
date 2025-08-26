package authctr

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/res"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httputil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type reqAuthLogout struct {
	ctx context.Context

	BodyRefreshToken string `json:"refreshToken" validate:"required"`
}

func (r *reqAuthLogout) Bind(c echo.Context) error {
	ctx, span := monitoring.Tracer().Start(r.ctx, "http.req.auth.logout")
	defer span.End()

	errMsgs := make(res.OLDVerboseMetaMsgs, 0)

	validationErr, err := httputil.BindOld(c, r, map[string]string{
		"refreshToken": "string",
	})
	if err != nil {
		otel.SpanLogError(span, err,
			otel.WithErrorLog(ctx),
			otel.WithMessage("failed to bind"),
		)
		return err
	}
	errMsgs.AppendDomMap(validationErr)

	validationErr, err = httputil.ValidateStruct(r, map[string]string{
		"BodyRefreshToken": "refreshToken",
	})
	if err != nil {
		otel.SpanLogError(span, err,
			otel.WithErrorLog(ctx),
			otel.WithMessage("unhandled validator error"),
		)
		return err
	}
	errMsgs.AppendDomMap(validationErr)

	if len(errMsgs) > 0 {
		return res.OLDNewErrorUnprocessable(errMsgs)
	}

	return nil
}

func (r *reqAuthLogout) Context() context.Context {
	return r.ctx
}

func (r *reqAuthLogout) RefreshToken() string {
	return r.BodyRefreshToken
}

func Logout(srv *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := monitoring.Tracer().Start(c.Request().Context(), "http.ctr.auth.logout")
		defer span.End()

		i := &reqAuthLogout{
			ctx: ctx,
		}

		err := i.Bind(c)
		if err != nil {
			if unErr, ok := err.(*res.OLDUnprocessableErrResponse); ok {
				return c.JSON(http.StatusUnprocessableEntity, unErr)
			}

			otel.SpanLogError(span, err,
				otel.WithErrorLog(ctx),
				otel.WithMessage("binding request error"),
			)
			return c.JSON(http.StatusInternalServerError, res.OLDNewErrorGeneric())
		}

		err = srv.Logout(i)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidToken) {
				return c.JSON(
					http.StatusUnauthorized,
					res.OLDNewError("Invalid refresh token", errcode.AuthInvalidCredentials),
				)
			}
			if errors.Is(err, auth.ErrTokenExpired) {
				return c.JSON(
					http.StatusUnauthorized,
					res.OLDNewError("Refresh token expired", errcode.AuthInvalidCredentials),
				)
			}

			otel.SpanLogError(span, err,
				otel.WithErrorLog(ctx),
				otel.WithMessage("error caught in service"),
			)
			return c.JSON(http.StatusInternalServerError, res.OLDNewErrorGeneric())
		}

		return c.NoContent(http.StatusNoContent)
	}
}
