package imig

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"komdigi-immigration/internal/app/core/logger/activitylog"
	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/core/util/timeutil"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/domain/actlog"
	"komdigi-immigration/internal/app/port"
)

type reqClearanceCheck interface {
	Context() context.Context
	Actor() domain.Userinfo

	LastName() string
	DateOfBirth() time.Time
	PassportNumber() string
	CountryCode() domain.CountryCode
	IMEI() domain.IMEI
}

func (srv *Imig) ClearanceCheck(
	req reqClearanceCheck,
) (domain.ImmigrationClearance, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), srv.sName("clearance-check"))
	defer span.End()

	actor := req.Actor()
	lastName := req.LastName()
	dateOfBirth := req.DateOfBirth()
	passportNumber := req.PassportNumber()
	countryCode := req.CountryCode()
	imei := req.IMEI()

	monitoring.Logger().InfoContext(
		ctx, "input",
		slog.Any("actor", actor),
		slog.String("lastName", lastName),
		slog.String("dateOfBirth", dateOfBirth.Format(timeutil.FormatDate)),
		slog.String("passportNumber", passportNumber),
		slog.String("countryCode", countryCode.String()),
		slog.String("imei", imei.String()),
	)

	emptyVal := domain.ImmigrationClearance{}
	if !countryCode.IsValid() {
		return emptyVal, ErrCountryCodeInvalid
	}

	logger := newClearanceLogger(
		srv.db,
		srv.icLogRepo,
		srv.activityLogger.New(actlog.ActionClearanceCheck),
		actor,
	)
	logger.IMEI(imei)
	logger.Passport(passportNumber, countryCode)

	defer func() {
		err := logger.Log(ctx)
		if err != nil {
			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to log service calling"),
			)
		}
	}()

	icRes, rawRes, resCode, err := srv.icService.Check(
		ctx, countryCode, passportNumber,
	)
	logger.RawResp(rawRes, resCode)
	if err != nil {
		if errors.Is(err, port.ErrClearanceNotFound) {
			logger.FinalErr("clearance information not found")
			return emptyVal, ErrClearanceInfoNotFound
		}

		logger.FinalErr("failed to check passport clearance to imigration service")
		return emptyVal, err
	}
	logger.ImmigrationClearance(icRes)

	if !strings.EqualFold(icRes.LastName, lastName) {
		logger.FinalErr("last name mismatch")
		return emptyVal, ErrLastNameMismatch
	}

	if !icRes.DateOfBirth.Equal(dateOfBirth) {
		logger.FinalErr("date of birth mismatch")
		return emptyVal, ErrDateOfBirthMismatch
	}

	logger.FinalOk()
	return icRes, nil
}

func newClearanceLogger(
	db *sql.DB,
	icLogRepo port.ImigrationClearanceLogRepo,
	activityLogger *activitylog.LogInstance,
	actor domain.Userinfo,
) *clearanceLogger {
	activityLogger.Actor(&actor)

	return &clearanceLogger{
		db:             db,
		icLogRepo:      icLogRepo,
		activityLogger: activityLogger,
		actor:          actor,
	}
}

type clearanceLogger struct {
	db             *sql.DB
	icLogRepo      port.ImigrationClearanceLogRepo
	activityLogger *activitylog.LogInstance
	actor          domain.Userinfo

	reqIMEI domain.IMEI

	resFirstName      *string
	resLastName       *string
	resGender         *domain.Gender
	resPlaceOfBirth   *string
	resDateOfBirth    *time.Time
	resPassportNumber *string
	resCountryCode    *domain.CountryCode
	resIsInIndonesia  *bool
	resIssuedAt       *time.Time
	resExpiredAt      *time.Time

	resRaw        []byte
	resStatusCode int64
}

func (cl *clearanceLogger) sName(s string) string {
	return "core.imig.clearance-logger." + s
}

func (cl *clearanceLogger) IMEI(imei domain.IMEI) {
	cl.reqIMEI = imei

	cl.activityLogger.AddMeta(actlog.MetaKeyReqIMEI, imei.String())
}

func (cl *clearanceLogger) Passport(
	passportNumber string, countryCode domain.CountryCode,
) {
	cl.activityLogger.AddMeta(actlog.MetaKeyResultPassportNumber, passportNumber)
	cl.activityLogger.AddMeta(actlog.MetaKeyResultCountryCode, countryCode.String())
}

func (cl *clearanceLogger) RawResp(rawRes []byte, statusCode int64) {
	cl.resRaw = rawRes
	cl.resStatusCode = statusCode

	cl.activityLogger.AddMeta(actlog.MetaKeyResultRaw, string(rawRes))
	cl.activityLogger.AddMeta(actlog.MetaKeyResultStatus, strconv.FormatInt(statusCode, 10))
}

func (cl *clearanceLogger) ImmigrationClearance(
	res domain.ImmigrationClearance,
) {
	cl.resFirstName = &res.FirstName
	cl.resLastName = &res.LastName
	cl.resGender = &res.Gender
	cl.resPlaceOfBirth = &res.PlaceOfBirth
	cl.resDateOfBirth = &res.DateOfBirth
	cl.resPassportNumber = &res.PassportNumber
	cl.resCountryCode = &res.CountryCode
	cl.resIsInIndonesia = &res.IsInIndonesia
	cl.resIssuedAt = &res.IssuedAt
	cl.resExpiredAt = &res.ExpiredAt

	cl.activityLogger.AddMeta(actlog.MetaKeyResultFirstName, res.FirstName)
	cl.activityLogger.AddMeta(actlog.MetaKeyResultLastName, res.LastName)
	cl.activityLogger.AddMeta(actlog.MetaKeyResultGender, res.Gender.String())
	cl.activityLogger.AddMeta(actlog.MetaKeyResultPlaceOfBirth, res.PlaceOfBirth)
	cl.activityLogger.AddMeta(actlog.MetaKeyResultDateOfBirth, res.DateOfBirth.Format(timeutil.FormatDate))
	cl.activityLogger.AddMeta(actlog.MetaKeyResultPassportNumber, res.PassportNumber)
	cl.activityLogger.AddMeta(actlog.MetaKeyResultCountryCode, res.CountryCode.String())
	cl.activityLogger.AddMeta(actlog.MetaKeyResultIsInIndonesia, strconv.FormatBool(res.IsInIndonesia))
	cl.activityLogger.AddMeta(actlog.MetaKeyResultIssuedAt, res.IssuedAt.Format(timeutil.FormatDate))
	cl.activityLogger.AddMeta(actlog.MetaKeyResultExpiredAt, res.ExpiredAt.Format(timeutil.FormatDate))
}

func (cl *clearanceLogger) FinalErr(errMsg string) {
	cl.activityLogger.AddMeta(actlog.MetaKeyFinalResponse, fmt.Sprintf("error: %s", errMsg))
}

func (cl *clearanceLogger) FinalOk() { cl.activityLogger.AddMeta(actlog.MetaKeyFinalResponse, "ok") }

func (cl *clearanceLogger) Log(ctx context.Context) error {
	ctx, span := monitoring.Tracer().Start(ctx, cl.sName("log-service-calling"))
	defer span.End()

	err := cl.icLogRepo.Log(
		ctx,
		cl.db,
		domain.NewImmigrationClearanceLogData(
			cl.resRaw,
			cl.resStatusCode,
			cl.resFirstName,
			cl.resLastName,
			cl.resGender,
			cl.resPlaceOfBirth,
			cl.resDateOfBirth,
			cl.resPassportNumber,
			cl.resCountryCode,
			cl.resIsInIndonesia,
			cl.resIssuedAt,
			cl.resExpiredAt,
			&cl.actor.User.User.ID,
			cl.reqIMEI,
		),
	)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to log service calling"),
		)
		return err
	}

	err = cl.activityLogger.Log(ctx)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to log activity"),
		)
		return err
	}

	return nil
}
