package actlogctr

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"komdigi-immigration/internal/app/adapter/in/http/req"
	"komdigi-immigration/internal/app/adapter/in/http/res"
	"komdigi-immigration/internal/app/core/logger"
	"komdigi-immigration/internal/app/core/util/errs"
	"komdigi-immigration/internal/app/core/util/httpfmt"
	"komdigi-immigration/internal/app/core/util/httpfmt/code/invalid"
	"komdigi-immigration/internal/app/core/util/httpfmt/rules"
	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/domain/actlog"
	"komdigi-immigration/internal/app/domain/errcode"

	"github.com/labstack/echo/v5"
)

func GetAll(srv *logger.Logger) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, span := monitoring.Tracer().Start(
			c.Request().Context(), "http.route.activity-log.get-all",
		)
		defer span.End()

		actor, err := req.AuthUser(ctx)
		if err != nil {
			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to get authenticated user"),
			)
			return c.JSON(http.StatusInternalServerError, httpfmt.NewErrorGeneric(err, httpfmt.WithTraceID(span)))
		}

		r := &reqGetAll{ctx: ctx, actor: actor}
		err = r.bind(c)
		if err != nil {
			return httpfmt.HandleUnprocessable(c, ctx, span, err)
		}

		logs, pg, err := srv.GetAll(r)
		if err != nil {
			if errs.Is(err, logger.ErrPermissionDenied) {
				return c.JSON(http.StatusForbidden, httpfmt.NewError(
					errcode.AuthPermissionDenied,
					"Permission denied",
					httpfmt.WithTraceID(span),
				))
			}

			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to get activity logs from core"),
			)
			return c.JSON(
				http.StatusInternalServerError,
				httpfmt.NewErrorGeneric(err, httpfmt.WithTraceID(span)),
			)
		}

		monitoring.Logger().DebugContext(
			ctx, "get activity response",
			slog.Any("pagination", pg),
		)

		return c.JSON(http.StatusOK, res.ActivityLogs(logs, pg))
	}
}

type reqGetAll struct {
	ctx   context.Context
	actor *domain.Userinfo

	QueryPage        string `query:"page"`
	QueryPerPage     string `query:"per_page"`
	QueryActionFlags string `query:"actions"`
	QueryUserIDFlags string `query:"user_ids"`
	QueryCreatedFrom string `query:"created_from"`
	QueryCreatedTo   string `query:"created_to"`

	page        int64
	perPage     int64
	actionFlags []actlog.Action
	userIDFlags []string
	createdFrom time.Time
	createdTo   time.Time
}

func (r *reqGetAll) bind(c *echo.Context) error {
	ctx, span := monitoring.Tracer().Start(
		r.ctx, "http.route.client-credential.get-all.bind",
	)
	defer span.End()

	uerr := httpfmt.NewUnprocessableErr(httpfmt.WithTraceID(span))

	err := httpfmt.Bind(c, uerr, r, map[string]invalid.Code{
		"page":         invalid.ShouldNumber,
		"per_page":     invalid.ShouldNumber,
		"actions":      invalid.ShouldEnum,
		"user_ids":     invalid.ShouldString,
		"created_from": invalid.ShouldDate,
		"created_to":   invalid.ShouldDate,
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
			"actions", r.QueryActionFlags,
		): {rules.EnumFlag(actlog.Action("").Options())},

		httpfmt.NewValidationInput(
			"created_from", r.QueryCreatedFrom,
		): {rules.Datetime()},
		httpfmt.NewValidationInput(
			"created_to", r.QueryCreatedTo,
		): {rules.Datetime()},
	})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to validate request"),
		)
		return err
	}

	page := func(raw string) int64 {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return 0
		}

		page, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return 0
		}
		return page
	}(r.QueryPage)

	perPage := func(raw string) int64 {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return 0
		}

		perPage, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return 0
		}
		return perPage
	}(r.QueryPerPage)

	actionFlags := func(raw string) []actlog.Action {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return nil
		}

		raws := strings.Split(trimmed, ",")
		actions := make([]actlog.Action, len(raws))
		for i, raw := range raws {
			actions[i] = actlog.Action(raw)
		}
		return actions
	}(r.QueryActionFlags)

	userIDFlags := func(raw string) []string {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return nil
		}

		return strings.Split(trimmed, ",")
	}(r.QueryUserIDFlags)

	createdFrom := func(raw string) time.Time {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return time.Time{}
		}

		res, err := time.Parse(time.RFC3339, trimmed)
		if err != nil {
			return time.Time{}
		}
		return res
	}(r.QueryCreatedFrom)

	createdTo := func(raw string) time.Time {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return time.Time{}
		}

		res, err := time.Parse(time.RFC3339, trimmed)
		if err != nil {
			return time.Time{}
		}
		return res
	}(r.QueryCreatedTo)

	if uerr.IsError() {
		return uerr
	}

	r.page = page
	r.perPage = perPage
	r.actionFlags = actionFlags
	r.userIDFlags = userIDFlags
	r.createdFrom = createdFrom
	r.createdTo = createdTo
	return nil
}

func (r *reqGetAll) ActionFlags() []actlog.Action { return r.actionFlags }
func (r *reqGetAll) Actor() domain.Userinfo       { return *r.actor }
func (r *reqGetAll) Context() context.Context     { return r.ctx }
func (r *reqGetAll) CreatedFrom() time.Time       { return r.createdFrom }
func (r *reqGetAll) CreatedTo() time.Time         { return r.createdTo }
func (r *reqGetAll) Page() int64                  { return r.page }
func (r *reqGetAll) PerPage() int64               { return r.perPage }
func (r *reqGetAll) UserIDFlags() []string        { return r.userIDFlags }
