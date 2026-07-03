package authctr

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"FaisalBudiono/go-boilerplate/internal/app/adapter/in/http/res"
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

func LoginWithEmail(srv *auth.Auth) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, span := monitoring.Tracer().Start(c.Request().Context(), "http.route.auth.login-with-email")
		defer span.End()

		r := &reqLoginWithEmail{ctx: ctx}

		err := r.bind(c)
		if err != nil {
			return httpfmt.HandleUnprocessable(c, ctx, span, err)
		}

		token, err := srv.LoginWithEmail(r)
		if err != nil {
			if errs.Is(err, auth.ErrInvalidCredentials) {
				monitoring.Logger().DebugContext(
					ctx, "invalid credentials",
					slog.Any("error", err),
				)
				return c.JSON(http.StatusUnauthorized, httpfmt.NewError(
					errcode.AuthInvalidCredentials,
					"invalid email or password",
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

		return c.JSON(http.StatusOK, res.TokenPair(token))
	}
}

type reqLoginWithEmail struct {
	ctx context.Context

	BodyEmail    string `json:"email"`
	BodyPassword string `json:"password"`

	email    string
	password string
}

func (r *reqLoginWithEmail) Context() context.Context { return r.ctx }
func (r *reqLoginWithEmail) Email() string            { return r.email }
func (r *reqLoginWithEmail) Password() string         { return r.password }

func (r *reqLoginWithEmail) bind(c *echo.Context) error {
	ctx, span := monitoring.Tracer().Start(r.ctx, "http.route.auth.login-with-email.bind")
	defer span.End()

	uerr := httpfmt.NewUnprocessableErr(httpfmt.WithTraceID(span))

	err := httpfmt.Bind(c, uerr, r, map[string]invalid.Code{
		"email":    invalid.ShouldString,
		"password": invalid.ShouldString,
	})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to bind request"),
		)
		return err
	}

	err = httpfmt.Validate(ctx, uerr, []httpfmt.RuleReq{
		httpfmt.NewRuleReq(
			"email", r.BodyEmail,
			rules.Required[string](),
		),
		httpfmt.NewRuleReq(
			"password", r.BodyPassword,
			rules.Required[string](),
		),
	})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to validate request"),
		)
		return err
	}

	email := strings.TrimSpace(r.BodyEmail)

	if uerr.IsError() {
		return uerr
	}

	r.email = email
	r.password = r.BodyPassword
	return nil
}
