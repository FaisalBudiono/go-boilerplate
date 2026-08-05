package userctr

import (
	"context"
	"net/http"
	"strings"

	"FaisalBudiono/go-boilerplate/internal/app/adapter/in/http/req"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/in/http/res"
	"FaisalBudiono/go-boilerplate/internal/app/core/user"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/errs"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt/code/invalid"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt/rules"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/mon"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"

	"github.com/labstack/echo/v5"
)

func Create(srv *user.User) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, span := mon.Tracer().Start(
			c.Request().Context(), "http.route.user.create",
		)
		defer span.End()

		actor, err := req.AuthUser(ctx)
		if err != nil {
			mon.Logger().DebugContext(
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

	BodyEmail    string   `json:"email"`
	BodyName     string   `json:"name"`
	BodyPassword string   `json:"password"`
	BodyRoles    []string `json:"roles"`

	email string
	name  string
	pass  string
	roles []domain.Role
}

func (r *reqCreate) bind(c *echo.Context) error {
	ctx, span := mon.Tracer().Start(
		r.ctx, "http.route.user.create.bind",
	)
	defer span.End()

	uerr := httpfmt.NewUnprocessableErr(httpfmt.WithTraceID(span))

	err := httpfmt.Bind(c, uerr, r, map[string]invalid.Code{
		"email":    invalid.ShouldString,
		"name":     invalid.ShouldString,
		"password": invalid.ShouldString,
		"roles":    invalid.ShouldArray,
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
			rules.Required[string](), rules.Email(), rules.MaxString(255),
		),
		httpfmt.NewRuleReq(
			"name", r.BodyName,
			rules.Required[string](), rules.MaxString(255),
		),
		httpfmt.NewRuleReq(
			"password", r.BodyPassword,
			rules.Required[string](),
		),
		httpfmt.NewRuleReq(
			"roles", r.BodyRoles,
			rules.SliceInside[string](rules.Enum(domain.Role("").Options())),
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
	name := strings.TrimSpace(r.BodyName)
	pass := r.BodyPassword
	roles := func(raws []string) []domain.Role {
		res := make([]domain.Role, len(raws))
		for i, raw := range raws {
			res[i] = domain.Role(raw)
		}
		return res
	}(r.BodyRoles)

	if uerr.IsError() {
		return uerr
	}

	r.email = email
	r.name = name
	r.pass = pass
	r.roles = roles
	return nil
}

func (r *reqCreate) Actor() domain.Userinfo   { return *r.actor }
func (r *reqCreate) Context() context.Context { return r.ctx }
func (r *reqCreate) Email() string            { return r.email }
func (r *reqCreate) Name() string             { return r.name }
func (r *reqCreate) Password() string         { return r.pass }
func (r *reqCreate) Roles() []domain.Role     { return r.roles }
