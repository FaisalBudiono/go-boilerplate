package userctr

import (
	"context"
	"net/http"

	"FaisalBudiono/go-boilerplate/internal/app/adapter/in/http/req"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/in/http/res"
	"FaisalBudiono/go-boilerplate/internal/app/core/user"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/errs"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"

	"github.com/labstack/echo/v5"
)

func Create(srv *user.User) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, span := monitoring.Tracer().Start(
			c.Request().Context(), "http.route.user.create",
		)
		defer span.End()

		actor, err := req.AuthUser(ctx)
		if err != nil {
			monitoring.Logger().DebugContext(
				ctx, "failed to get authenticated user",
			)
			return c.JSON(
				http.StatusInternalServerError,
				httpfmt.NewErrorGeneric(err, httpfmt.WithTraceID(span)),
			)
		}

		r := &reqCreate{ctx: ctx, actor: actor}
		err = r.bind(c)
		if err != nil {
			return httpfmt.HandleUnprocessable(c, ctx, span, err)
		}

		u, err := srv.Create(r)
		if err != nil {
			if errs.Is(err, user.ErrPermissionDenied) {
				return c.JSON(http.StatusForbidden, httpfmt.NewError(
					errcode.AuthPermissionDenied,
					"Permission denied",
					httpfmt.WithTraceID(span),
				))
			}

			if errs.Is(err, user.ErrEmailDuplicated) {
				return c.JSON(http.StatusConflict, httpfmt.NewError(
					errcode.UserEmailExists,
					"Email already exists",
					httpfmt.WithTraceID(span),
				))
			}

			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to create user from core"),
			)
			return c.JSON(
				http.StatusInternalServerError,
				httpfmt.NewErrorGeneric(err, httpfmt.WithTraceID(span)),
			)
		}

		return c.JSON(http.StatusCreated, res.User(u))
	}
}

type reqCreate struct {
	ctx   context.Context
	actor *domain.Userinfo

	email string
	name  string
	pass  string
	roles []domain.Role
}

func (r *reqCreate) bind(c *echo.Context) error {
	panic("unimplemented")
}

func (r *reqCreate) Actor() domain.Userinfo   { return *r.actor }
func (r *reqCreate) Context() context.Context { return r.ctx }
func (r *reqCreate) Email() string            { return r.email }
func (r *reqCreate) Name() string             { return r.email }
func (r *reqCreate) Password() string         { return r.pass }
func (r *reqCreate) Roles() []domain.Role     { return r.roles }
