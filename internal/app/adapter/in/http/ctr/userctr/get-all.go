package userctr

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"komdigi-immigration/internal/app/adapter/in/http/req"
	"komdigi-immigration/internal/app/adapter/in/http/res"
	"komdigi-immigration/internal/app/core/user"
	"komdigi-immigration/internal/app/core/util/errs"
	"komdigi-immigration/internal/app/core/util/httpfmt"
	"komdigi-immigration/internal/app/core/util/httpfmt/code/invalid"
	"komdigi-immigration/internal/app/core/util/httpfmt/rules"
	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/domain/errcode"

	"github.com/labstack/echo/v5"
)

func GetAll(srv *user.User) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, span := monitoring.Tracer().Start(c.Request().Context(), "http.route.user.get-all")
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

		r := &reqGetAll{ctx: ctx, actor: *actor}

		err = r.bind(c)
		if err != nil {
			return httpfmt.HandleUnprocessable(c, ctx, span, err)
		}

		users, pg, err := srv.GetAll(r)
		if err != nil {
			if errs.Is(err, user.ErrPermissionDenied) {
				return c.JSON(http.StatusForbidden, httpfmt.NewError(
					errcode.AuthPermissionDenied,
					"Permission denied",
					httpfmt.WithTraceID(span),
				))
			}

			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to get users"),
			)
			return c.JSON(
				http.StatusInternalServerError,
				httpfmt.NewErrorGeneric(err, httpfmt.WithTraceID(span)),
			)
		}

		return c.JSON(http.StatusOK, res.Users(users, pg))
	}
}

type reqGetAll struct {
	ctx context.Context

	QueryPage     string `query:"page"`
	QueryPerPage  string `query:"per_page"`
	QueryRoleFlag string `query:"roles"`

	actor     domain.Userinfo
	page      *int64
	perPage   *int64
	roleFlags []domain.Role
}

func (r *reqGetAll) bind(c *echo.Context) error {
	ctx, span := monitoring.Tracer().Start(r.ctx, "http.route.user.get-all.bind")
	defer span.End()

	uerr := httpfmt.NewUnprocessableErr(httpfmt.WithTraceID(span))

	err := httpfmt.Bind(c, uerr, r, map[string]invalid.Code{
		"page":     invalid.ShouldNumber,
		"per_page": invalid.ShouldNumber,
		"roles":    invalid.ShouldEnum,
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
			"page", r.QueryPage,
		): {rules.StrToInt()},
		httpfmt.NewValidationInput(
			"per_page", r.QueryPerPage,
		): {rules.StrToInt()},
		httpfmt.NewValidationInput(
			"roles", r.QueryRoleFlag,
		): {rules.EnumFlag(domain.Role("").Options())},
	})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to validate request"),
		)
		return err
	}

	page := func(raw string) *int64 {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return nil
		}
		page, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return nil
		}
		return &page
	}(r.QueryPage)

	perPage := func(raw string) *int64 {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return nil
		}
		perPage, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return nil
		}
		return &perPage
	}(r.QueryPerPage)

	roleFlags := func(raw string) []domain.Role {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return nil
		}
		raws := strings.Split(trimmed, ",")
		roles := make([]domain.Role, len(raws))
		for i, raw := range raws {
			roles[i] = domain.Role(raw)
		}
		return roles
	}(r.QueryRoleFlag)

	if uerr.IsError() {
		return uerr
	}

	r.page = page
	r.perPage = perPage
	r.roleFlags = roleFlags

	return nil
}

func (r *reqGetAll) Context() context.Context { return r.ctx }
func (r *reqGetAll) Actor() domain.Userinfo   { return r.actor }
func (r *reqGetAll) Page() *int64             { return r.page }
func (r *reqGetAll) PerPage() *int64          { return r.perPage }
func (r *reqGetAll) RoleFlags() []domain.Role { return r.roleFlags }
