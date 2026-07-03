package authctr

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

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

func LoginWithClientID(srv *auth.Auth) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, span := monitoring.Tracer().Start(
			c.Request().Context(), "http.route.auth.login-with-client-id",
		)
		defer span.End()

		r := &reqLoginWithClientID{ctx: ctx}
		err := r.bind(c)
		if err != nil {
			return httpfmt.HandleUnprocessable(c, ctx, span, err)
		}

		tp, err := srv.LoginWithClientID(r)
		if err != nil {
			if errs.Is(err, auth.ErrInvalidCredentials) {
				monitoring.Logger().DebugContext(
					ctx, "invalid credentials",
					slog.Any("error", err),
				)
				return c.JSON(http.StatusUnauthorized, httpfmt.NewError(
					errcode.AuthInvalidCredentials,
					"invalid clientID or clientSecret",
					httpfmt.WithTraceID(span),
				))
			}

			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to login from core"),
			)
			return c.JSON(
				http.StatusInternalServerError,
				httpfmt.NewErrorGeneric(err, httpfmt.WithTraceID(span)),
			)
		}

		return c.JSON(http.StatusOK, res.TokenPair(tp))
	}
}

type reqLoginWithClientID struct {
	ctx context.Context

	BodyClientID     string `json:"clientID"`
	BodyClientSecret string `json:"clientSecret"`

	clientID     string
	clientSecret string
}

func (r *reqLoginWithClientID) bind(c *echo.Context) error {
	ctx, span := monitoring.Tracer().Start(
		r.ctx, "http.route.auth.login-with-client-id.bind",
	)
	defer span.End()

	uerr := httpfmt.NewUnprocessableErr(httpfmt.WithTraceID(span))

	err := httpfmt.Bind(c, uerr, r, map[string]invalid.Code{
		"clientID":     invalid.ShouldString,
		"clientSecret": invalid.ShouldString,
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
			"clientID", r.BodyClientID,
		): {rules.Required[string]()},
		httpfmt.NewValidationInput(
			"clientSecret", r.BodyClientSecret,
		): {rules.Required[string]()},
	})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to validate request"),
		)
		return err
	}

	clientID := strings.TrimSpace(r.BodyClientID)
	clientSecret := strings.TrimSpace(r.BodyClientSecret)

	if uerr.IsError() {
		return uerr
	}

	r.clientID = clientID
	r.clientSecret = clientSecret
	return nil
}

func (r *reqLoginWithClientID) Context() context.Context { return r.ctx }
func (r *reqLoginWithClientID) ClientID() string         { return r.clientID }
func (r *reqLoginWithClientID) ClientSecret() string     { return r.clientSecret }
