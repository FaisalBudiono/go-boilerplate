package authctr

import (
	"context"
	"net/http"

	"komdigi-immigration/internal/app/adapter/in/http/res"
	"komdigi-immigration/internal/app/core/auth"
	"komdigi-immigration/internal/app/core/util/errs"
	"komdigi-immigration/internal/app/core/util/httpfmt"
	"komdigi-immigration/internal/app/core/util/httpfmt/code/invalid"
	"komdigi-immigration/internal/app/core/util/httpfmt/rules"
	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/domain/errcode"

	"github.com/labstack/echo/v5"
)

func RefreshToken(srv *auth.Auth) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, span := monitoring.Tracer().Start(c.Request().Context(), "http.route.auth.refresh-token")
		defer span.End()

		r := &reqRefreshToken{ctx: ctx}

		err := r.bind(c)
		if err != nil {
			return httpfmt.HandleUnprocessable(c, ctx, span, err)
		}

		token, err := srv.RefreshToken(r)
		if err != nil {
			if errs.Is(err, auth.ErrTokenExpired) {
				return c.JSON(http.StatusUnauthorized, httpfmt.NewError(
					errcode.AuthTokenExpired,
					"Token expired",
					httpfmt.WithTraceID(span),
				))
			}

			if errs.Is(err, auth.ErrTokenInvalid, auth.ErrInvalidCredentials) {
				return c.JSON(http.StatusUnauthorized, httpfmt.NewError(
					errcode.AuthTokenInvalid,
					"Invalid refresh token",
					httpfmt.WithTraceID(span),
				))
			}

			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed refreshToken from core"),
			)
			return c.JSON(
				http.StatusInternalServerError,
				httpfmt.NewErrorGeneric(err, httpfmt.WithTraceID(span)),
			)
		}

		return c.JSON(http.StatusOK, res.TokenPair(token))
	}
}

type reqRefreshToken struct {
	ctx context.Context

	BodyRefreshToken string `json:"refreshToken"`

	refreshToken string
}

func (r *reqRefreshToken) bind(c *echo.Context) error {
	ctx, span := monitoring.Tracer().Start(r.ctx, "http.route.auth.refresh-token.bind")
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

func (r *reqRefreshToken) Context() context.Context { return r.ctx }
func (r *reqRefreshToken) RefreshToken() string     { return r.refreshToken }
