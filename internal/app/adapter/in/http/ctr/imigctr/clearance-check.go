package imigctr

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"komdigi-immigration/internal/app/adapter/in/http/req"
	"komdigi-immigration/internal/app/adapter/in/http/res"
	"komdigi-immigration/internal/app/core/imig"
	"komdigi-immigration/internal/app/core/util/errs"
	"komdigi-immigration/internal/app/core/util/httpfmt"
	"komdigi-immigration/internal/app/core/util/httpfmt/code/invalid"
	"komdigi-immigration/internal/app/core/util/httpfmt/rules"
	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/core/util/timeutil"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/domain/errcode"

	"github.com/labstack/echo/v5"
)

func ClearanceCheck(srv *imig.Imig) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, span := monitoring.Tracer().Start(
			c.Request().Context(), "http.route.immigration.clearance-check",
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

		r := &reqClearanceCheck{ctx: ctx, actor: actor}
		err = r.bind(c)
		if err != nil {
			return httpfmt.HandleUnprocessable(c, ctx, span, err)
		}

		ic, err := srv.ClearanceCheck(r)
		if err != nil {
			if errs.Is(err, imig.ErrClearanceInfoNotFound) {
				monitoring.Logger().DebugContext(
					ctx, "clearance information not found",
					slog.Any("error", err),
				)
				return c.JSON(http.StatusNotFound, httpfmt.NewError(
					errcode.NotFound,
					"Clearance information not found",
					httpfmt.WithTraceID(span),
				))
			}

			if errs.Is(err, imig.ErrLastNameMismatch, imig.ErrDateOfBirthMismatch) {
				monitoring.Logger().DebugContext(
					ctx, "clearance information mismatch",
					slog.Any("error", err),
				)
				return c.JSON(http.StatusConflict, httpfmt.NewError(
					errcode.ImigClearanceInvalidValidation,
					"Clearance information mismatch",
					httpfmt.WithTraceID(span),
				))
			}

			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to do clearance check from core"),
			)
			return c.JSON(
				http.StatusInternalServerError,
				httpfmt.NewErrorGeneric(err, httpfmt.WithTraceID(span)),
			)
		}

		return c.JSON(http.StatusOK, res.ImmigrationClearance(ic))
	}
}

type reqClearanceCheck struct {
	ctx   context.Context
	actor *domain.Userinfo

	BodyLastName    string `json:"lastName"`
	BodyDateOfBirth string `json:"dateOfBirth"`
	BodyPassportNum string `json:"passportNumber"`
	BodyCountryCode string `json:"countryCode"`
	BodyIMEI        string `json:"imei"`

	lastName    string
	dateOfBirth time.Time
	passportNum string
	countryCode domain.CountryCode
	imei        domain.IMEI
}

func (r *reqClearanceCheck) bind(c *echo.Context) error {
	ctx, span := monitoring.Tracer().Start(
		r.ctx, "http.route.immigration.clearance-check.bind",
	)
	defer span.End()

	uerr := httpfmt.NewUnprocessableErr(httpfmt.WithTraceID(span))

	err := httpfmt.Bind(c, uerr, r, map[string]invalid.Code{
		"lastName":       invalid.ShouldString,
		"dateOfBirth":    invalid.ShouldString,
		"passportNumber": invalid.ShouldString,
		"countryCode":    invalid.ShouldString,
		"imei":           invalid.ShouldString,
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
			"lastName", r.BodyLastName,
		): {rules.Required[string]()},
		httpfmt.NewValidationInput(
			"dateOfBirth", r.BodyDateOfBirth,
		): {rules.Required[string](), rules.Date()},

		httpfmt.NewValidationInput(
			"passportNumber", r.BodyPassportNum,
		): {rules.Required[string]()},
		httpfmt.NewValidationInput(
			"countryCode", r.BodyCountryCode,
		): {
			rules.Required[string](),
			rules.EnumFlag(domain.CountryCode("").Options()),
		},
		httpfmt.NewValidationInput(
			"imei", r.BodyIMEI,
		): {rules.Required[string](), rules.Luhn()},
	})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to validate request"),
		)
		return err
	}

	lastName := func(raw string) string {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return ""
		}
		return trimmed
	}(r.BodyLastName)

	dateOfBirth := func(raw string) time.Time {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return time.Time{}
		}
		res, err := time.Parse(timeutil.FormatDate, trimmed)
		if err != nil {
			return time.Time{}
		}
		return res.UTC()
	}(r.BodyDateOfBirth)

	passportNumber := func(raw string) string {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return ""
		}
		return trimmed
	}(r.BodyPassportNum)

	countryCode := func(raw string) domain.CountryCode {
		emptyVal := domain.CountryCode("")
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return emptyVal
		}
		res := domain.CountryCode(trimmed)
		if !res.IsValid() {
			return emptyVal
		}
		return res
	}(r.BodyCountryCode)

	imei := func(raw string) domain.IMEI {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return ""
		}
		return domain.IMEI(trimmed)
	}(r.BodyIMEI)

	if uerr.IsError() {
		return uerr
	}

	r.lastName = lastName
	r.dateOfBirth = dateOfBirth
	r.passportNum = passportNumber
	r.countryCode = countryCode
	r.imei = imei

	return nil
}

func (r *reqClearanceCheck) Actor() domain.Userinfo          { return *r.actor }
func (r *reqClearanceCheck) Context() context.Context        { return r.ctx }
func (r *reqClearanceCheck) CountryCode() domain.CountryCode { return r.countryCode }
func (r *reqClearanceCheck) DateOfBirth() time.Time          { return r.dateOfBirth }
func (r *reqClearanceCheck) IMEI() domain.IMEI               { return r.imei }
func (r *reqClearanceCheck) LastName() string                { return r.lastName }
func (r *reqClearanceCheck) PassportNumber() string          { return r.passportNum }
