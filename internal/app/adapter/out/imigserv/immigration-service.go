package imigserv

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"komdigi-immigration/internal/app/adapter/hardcode/passport"
	"komdigi-immigration/internal/app/core/util/httpfmt"
	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/core/util/timeutil"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/domain/errcode"
	"komdigi-immigration/internal/app/port"
)

func New(ctx context.Context) (*ImmigrationService, error) {
	hc := passport.New()
	res, err := hc.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	return &ImmigrationService{
		doms: res,
	}, nil
}

type ImmigrationService struct {
	doms []domain.ImmigrationClearance
}

func (is *ImmigrationService) Check(
	ctx context.Context,
	countryCode domain.CountryCode,
	passportNumber string,
) (domain.ImmigrationClearance, []byte, int64, error) {
	ctx, span := monitoring.Tracer().Start(ctx, is.sName("check"))
	defer span.End()

	monitoring.Logger().InfoContext(
		ctx, "checking passport",
		slog.String("passportNumber", passportNumber),
		slog.String("countryCode", string(countryCode)),
	)

	emptyVal := domain.ImmigrationClearance{}

	for _, d := range is.doms {
		isFound := strings.EqualFold(d.PassportNumber, passportNumber) &&
			strings.EqualFold(string(d.CountryCode), string(countryCode))

		if !isFound {
			continue
		}

		monitoring.Logger().DebugContext(
			ctx, "found", slog.Any("dom", d),
		)

		raw := newRawDummy(d)
		rawResp, err := json.Marshal(raw)
		if err != nil {
			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to marshal raw dummy for success"),
			)

			return emptyVal, []byte(err.Error()), domain.StatusCodeErrInternal, err
		}

		return d, rawResp, http.StatusOK, nil
	}

	raw := httpfmt.NewError(errcode.NotFound, "passport information not found", httpfmt.WithTraceID(span))
	rawResp, err := json.Marshal(raw)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to marshal raw dummy for not found"),
		)
		return emptyVal, []byte(err.Error()), domain.StatusCodeErrInternal, err
	}

	return emptyVal, rawResp, http.StatusNotFound, port.ErrClearanceNotFound
}

func (is *ImmigrationService) sName(s string) string {
	return "adapter.out.hardcode.passport.service." + s
}

type rawDummy struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Gender    string `json:"gender"`

	PlaceOfBirth string `json:"placeOfBirth"`
	DateOfBirth  string `json:"dateOfBirth"`

	PassportNumber string `json:"passportNumber"`
	CountryCode    string `json:"countryCode"`

	IssuedAt  string `json:"issuedAt"`
	ExpiredAt string `json:"expiredAt"`
}

func newRawDummy(d domain.ImmigrationClearance) rawDummy {
	return rawDummy{
		FirstName: d.FirstName,
		LastName:  d.LastName,
		Gender:    d.Gender.String(),

		PlaceOfBirth: d.PlaceOfBirth,
		DateOfBirth:  d.DateOfBirth.Format(timeutil.FormatDate),

		PassportNumber: d.PassportNumber,
		CountryCode:    d.CountryCode.String(),

		IssuedAt:  d.IssuedAt.Format(timeutil.FormatDate),
		ExpiredAt: d.ExpiredAt.Format(timeutil.FormatDate),
	}
}
