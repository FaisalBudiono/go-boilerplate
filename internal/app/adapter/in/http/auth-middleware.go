package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"FaisalBudiono/go-boilerplate/internal/app/adapter/in/http/req"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"

	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func AuthMiddleware(authCore *auth.Auth, opts ...authOption) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx, span := monitoring.Tracer().Start(c.Request().Context(), "http.middleware.auth")
			defer span.End()

			cfg := newAuthConfig()
			for _, opt := range opts {
				opt(cfg)
			}

			monitoring.Logger().DebugContext(
				ctx, "auth config",
				slog.Any("allowedRoles", cfg.allowedRoles),
				slog.Bool("isMust", cfg.isMust),
				slog.Any("cfg", cfg),
			)

			authHeader := c.Request().Header.Get("authorization")
			if authHeader == "" {
				monitoring.Logger().InfoContext(ctx, "missing authorization header")

				if !cfg.isMust {
					return next(c)
				}
				return c.JSON(
					http.StatusUnauthorized,
					httpfmt.NewError(
						errcode.AuthUnauthorized,
						"Authorization header required",
						httpfmt.WithTraceID(span),
					),
				)
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 {
				monitoring.Logger().InfoContext(
					ctx, "invalid authorization header format",
					slog.String("authHeader", authHeader),
				)

				return c.JSON(
					http.StatusUnauthorized,
					httpfmt.NewError(
						errcode.AuthUnauthorized,
						"Invalid authorization format. Expected: Bearer <token>",
						httpfmt.WithTraceID(span),
					),
				)
			}

			user, err := authCore.ParseUser(&httpParseUserReq{
				ctx:         ctx,
				accessToken: parts[1],
			})
			if err != nil {
				if errors.Is(err, auth.ErrTokenExpired) {
					monitoring.Logger().InfoContext(ctx, "token expired")

					return c.JSON(
						http.StatusUnauthorized,
						httpfmt.NewError(
							errcode.AuthTokenExpired,
							"Token has expired",
							httpfmt.WithTraceID(span),
						),
					)
				}

				if errors.Is(err, auth.ErrTokenInvalid) {
					monitoring.Logger().DebugContext(ctx, "invalid token")

					return c.JSON(
						http.StatusUnauthorized,
						httpfmt.NewError(
							errcode.AuthTokenInvalid,
							"Invalid token",
							httpfmt.WithTraceID(span),
						),
					)
				}

				otelutil.SpanLogError(
					span, err,
					otelutil.WithErrorLog(ctx),
					otelutil.WithMessage("authentication error"),
				)

				return c.JSON(
					http.StatusInternalServerError,
					httpfmt.NewErrorGeneric(err, httpfmt.WithTraceID(span)),
				)
			}

			roleStrings := make([]string, len(user.User.Roles))
			for i, role := range user.User.Roles {
				roleStrings[i] = string(role)
			}

			span.AddEvent("user authenticated", trace.WithAttributes(
				attribute.String("userID", user.Info.ID),
				attribute.String("email", user.User.User.Email),
				attribute.String("loginID", user.Info.LoginID),
				attribute.String("loginMethod", string(user.Info.LoginMethod)),
				attribute.StringSlice("roles", roleStrings),
			))
			monitoring.Logger().DebugContext(
				ctx, "user authenticated",
				slog.String("userID", user.Info.ID),
				slog.String("email", user.User.User.Email),
				slog.String("loginID", user.Info.LoginID),
				slog.String("loginMethod", string(user.Info.LoginMethod)),
				slog.Any("roles", roleStrings),
			)

			c.SetRequest(c.Request().WithContext(req.ContextWithUser(ctx, &user)))

			if len(cfg.allowedRoles) > 0 {
				if !user.User.HasRoles(cfg.allowedRoles...) {
					monitoring.Logger().WarnContext(ctx, "user is not allowed to access this resource")

					allowedRoles := make([]string, len(cfg.allowedRoles))
					for i, role := range cfg.allowedRoles {
						allowedRoles[i] = string(role)
					}

					return c.JSON(
						http.StatusForbidden,
						httpfmt.NewError(
							errcode.AuthUnauthorized,
							fmt.Sprintf("Access denied. Allowed roles are (%s)", strings.Join(allowedRoles, ",")),
							httpfmt.WithTraceID(span),
						),
					)
				}
			}

			return next(c)
		}
	}
}

type httpParseUserReq struct {
	ctx         context.Context
	accessToken string
}

func (r *httpParseUserReq) Context() context.Context {
	return r.ctx
}

func (r *httpParseUserReq) AccessToken() string {
	return r.accessToken
}

func newAuthConfig() *authConfig {
	return &authConfig{
		allowedRoles: []domain.Role{},
		isMust:       false,
	}
}

type authConfig struct {
	allowedRoles []domain.Role
	isMust       bool
}

type authOption func(*authConfig)

func WithRoles(roles ...domain.Role) authOption {
	return func(opts *authConfig) {
		opts.allowedRoles = roles
	}
}

func WithNeedAuth() authOption {
	return func(opts *authConfig) {
		opts.isMust = true
	}
}
