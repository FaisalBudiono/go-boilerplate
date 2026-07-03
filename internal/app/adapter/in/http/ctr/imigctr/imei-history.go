package imigctr

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"komdigi-immigration/internal/app/adapter/in/http/req"
	"komdigi-immigration/internal/app/adapter/in/http/res"
	"komdigi-immigration/internal/app/core/imig"
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

func IMEIHistory(srv *imig.Imig) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, span := monitoring.Tracer().Start(
			c.Request().Context(), "http.route.immigration.imei-history",
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

		r := &reqIMEIHistory{ctx: ctx, actor: actor}
		err = r.bind(c)
		if err != nil {
			return httpfmt.HandleUnprocessable(c, ctx, span, err)
		}

		ics, pg, err := srv.IMEIHistory(r)
		if err != nil {
			if errs.Is(err, imig.ErrPermissionDenied) {
				return c.JSON(http.StatusForbidden, httpfmt.NewError(
					errcode.AuthPermissionDenied,
					"Permission denied",
					httpfmt.WithTraceID(span),
				))
			}

			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to get imei history from core"),
			)
			return c.JSON(
				http.StatusInternalServerError,
				httpfmt.NewErrorGeneric(err, httpfmt.WithTraceID(span)),
			)
		}

		return c.JSON(http.StatusOK, res.PaginatedImmigrationClearanceLog(ics, pg))
	}
}

type reqIMEIHistory struct {
	ctx   context.Context
	actor *domain.Userinfo

	ParamIMEI    string `param:"imei"`
	QueryPage    string `query:"page"`
	QueryPerPage string `query:"per_page"`

	page    int64
	perPage int64
	imei    string
}

func (r *reqIMEIHistory) bind(c *echo.Context) error {
	ctx, span := monitoring.Tracer().Start(
		r.ctx, "http.route.immigration.imei-history.bind",
	)
	defer span.End()

	uerr := httpfmt.NewUnprocessableErr(httpfmt.WithTraceID(span))

	err := httpfmt.Bind(c, uerr, r, map[string]invalid.Code{
		"imei":     invalid.ShouldString,
		"page":     invalid.ShouldNumber,
		"per_page": invalid.ShouldNumber,
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
			"imei", r.ParamIMEI,
		): {rules.Required[string]()},

		httpfmt.NewValidationInput(
			"page", r.QueryPage,
		): {rules.StrToInt()},
		httpfmt.NewValidationInput(
			"per_page", r.QueryPerPage,
		): {rules.StrToInt()},
	})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to validate request"),
		)
		return err
	}

	imei := strings.TrimSpace(r.ParamIMEI)

	page := func(raw string) int64 {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return 0
		}
		res, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return 0
		}
		return res
	}(r.QueryPage)

	perPage := func(raw string) int64 {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return 0
		}
		res, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return 0
		}
		return res
	}(r.QueryPerPage)

	if uerr.IsError() {
		return uerr
	}

	r.imei = imei
	r.page = page
	r.perPage = perPage

	return nil
}

func (r *reqIMEIHistory) Actor() domain.Userinfo   { return *r.actor }
func (r *reqIMEIHistory) Context() context.Context { return r.ctx }
func (r *reqIMEIHistory) IMEI() string             { return r.imei }
func (r *reqIMEIHistory) Page() int64              { return r.page }
func (r *reqIMEIHistory) PerPage() int64           { return r.perPage }
