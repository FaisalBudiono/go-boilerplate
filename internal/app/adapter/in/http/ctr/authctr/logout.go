package authctr

import (
	"context"
	"net/http"

	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/errs"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt/code/invalid"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt/rules"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"

	"github.com/labstack/echo/v5"
)

func Logout(srv *auth.Auth) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, span := monitoring.Tracer().Start(c.Request().Context(), "http.route.auth.logout")
		defer span.End()

		r := &reqLogout{ctx: ctx}

		err := r.bind(c)
		if err != nil {
			return httpfmt.HandleUnprocessable(c, ctx, span, err)
		}

		err = srv.Logout(r)
		if err != nil {
			if errs.Is(err, auth.ErrInvalidCredentials) {
				return c.NoContent(http.StatusNoContent)
			}

			if errs.Is(err, auth.ErrTokenExpired) {
				return c.JSON(http.StatusUnauthorized, httpfmt.NewError(
					errcode.AuthTokenExpired,
					"Token expired",
					httpfmt.WithTraceID(span),
				))
			}

			if errs.Is(err, auth.ErrTokenInvalid) {
				return c.JSON(http.StatusUnauthorized, httpfmt.NewError(
					errcode.AuthTokenInvalid,
					"Invalid refresh token",
					httpfmt.WithTraceID(span),
				))
			}

			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed logout from core"),
			)
			return c.JSON(
				http.StatusInternalServerError,
				httpfmt.NewErrorGeneric(err, httpfmt.WithTraceID(span)),
			)
		}

		return c.NoContent(http.StatusNoContent)
	}
}

type reqLogout struct {
	ctx context.Context

	BodyRefreshToken string `json:"refreshToken"`

	refreshToken string
}

func (r *reqLogout) bind(c *echo.Context) error {
	ctx, span := monitoring.Tracer().Start(c.Request().Context(), "http.route.auth.logout.bind")
	defer span.End()

	uerr := httpfmt.NewUnprocessableErr(httpfmt.WithTraceID(span))

	err := httpfmt.Bind(c, uerr, r, map[string]invalid.Code{
		"refreshToken": invalid.ShouldString,
	})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to bind request"),
		)
		return err
	}

	err = httpfmt.Validate(ctx, uerr, map[httpfmt.ValidationInput][]httpfmt.Rule{
		httpfmt.NewValidationInput(
			"refreshToken", r.BodyRefreshToken,
		): {rules.Required[string]()},
	})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to validate request"),
		)
		return err
	}

	if uerr.IsError() {
		return uerr
	}

	r.refreshToken = r.BodyRefreshToken

	return nil
}

func (r *reqLogout) Context() context.Context { return r.ctx }
func (r *reqLogout) RefreshToken() string     { return r.refreshToken }
