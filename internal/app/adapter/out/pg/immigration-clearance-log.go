package pg

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/port"
)

func NewImmigrationClearanceLog() *ImmigrationClearanceLog {
	return &ImmigrationClearanceLog{}
}

type immigrationClearanceLogRes struct {
	id          string
	rawResponse string
	statusCode  int64

	firstName *string
	lastName  *string
	gender    *string

	placeOfBirth *string
	dateOfBirth  *time.Time

	passportNumber *string
	countryCode    *string

	isInIndonesia *bool

	issuedAt  *time.Time
	expiredAt *time.Time

	userID *string
	imei   string

	createdAt time.Time
}

type ImmigrationClearanceLog struct{}

func (repo *ImmigrationClearanceLog) GetIMEIHistoryPaginated(
	ctx context.Context, tx port.DBTX, imei string, page int64, perPage int64,
) ([]domain.ImmigrationClearanceLog, int64, error) {
	ctx, span := monitoring.Tracer().Start(ctx, repo.sName("get-imei-history-paginated"))
	defer span.End()

	monitoring.Logger().DebugContext(
		ctx, "input", slog.String("imei", imei),
		slog.Int64("page", page),
		slog.Int64("perPage", perPage),
	)

	conditions := []string{
		"icl.deleted_at IS NULL",
		"icl.status_code >= 200",
		"icl.status_code < 300",
		"icl.imei = $1",
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(
		`
		SELECT
			icl.id,
			icl.raw_response,
			icl.status_code,
			icl.first_name,
			icl.last_name,
			icl.gender,
			icl.place_of_birth,
			icl.date_of_birth,
			icl.passport_number,
			icl.country_code,
			icl.is_in_indonesia,
			icl.issued_at,
			icl.expired_at,
			icl.user_id,
			icl.imei,
			icl.created_at
		FROM
			immigration_clearance_logs AS icl
		WHERE
			%s
		LIMIT %d OFFSET %d
	`,
		strings.Join(conditions, " AND "),
		perPage,
		offset,
	)
	args := []any{imei}

	monitoring.Logger().DebugContext(
		ctx, "query", slog.String("query", query),
		slog.Any("args", args),
	)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to query immigration clearance logs by imei"),
		)
		return nil, 0, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to close rows"),
			)
		}
	}()

	res := make([]domain.ImmigrationClearanceLog, 0)
	for rows.Next() {
		var raw immigrationClearanceLogRes
		err = rows.Scan(
			&raw.id,
			&raw.rawResponse,
			&raw.statusCode,
			&raw.firstName,
			&raw.lastName,
			&raw.gender,
			&raw.placeOfBirth,
			&raw.dateOfBirth,
			&raw.passportNumber,
			&raw.countryCode,
			&raw.isInIndonesia,
			&raw.issuedAt,
			&raw.expiredAt,
			&raw.userID,
			&raw.imei,
			&raw.createdAt,
		)
		if err != nil {
			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to scan immigration clearance log"),
			)
			return nil, 0, err
		}

		var gender *domain.Gender
		if raw.gender != nil {
			gender = new(domain.Gender(*raw.gender))
		}

		var countryCode *domain.CountryCode
		if raw.countryCode != nil {
			countryCode = new(domain.CountryCode(*raw.countryCode))
		}

		res = append(res, domain.NewImmigrationClearanceLog(
			raw.id,
			[]byte(raw.rawResponse),
			raw.statusCode,
			raw.firstName,
			raw.lastName,
			gender,
			raw.placeOfBirth,
			raw.dateOfBirth,
			raw.passportNumber,
			countryCode,
			raw.isInIndonesia,
			raw.issuedAt,
			raw.expiredAt,
			raw.userID,
			domain.IMEI(raw.imei),
			raw.createdAt,
		))
	}

	countQuery := fmt.Sprintf(
		`
		SELECT
			COUNT(1)
		FROM
			immigration_clearance_logs AS icl
		WHERE
			%s
		`,
		strings.Join(conditions, " AND "),
	)

	monitoring.Logger().DebugContext(
		ctx, "count query", slog.String("query", query),
		slog.Any("args", args),
	)

	var total int64
	err = tx.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to counting immigration clearance logs by imei"),
		)
		return nil, 0, err
	}

	return res, total, nil
}

func (repo *ImmigrationClearanceLog) Log(
	ctx context.Context,
	tx port.DBTX,
	log domain.ImmigrationClearanceLogData,
) error {
	ctx, span := monitoring.Tracer().Start(ctx, repo.sName("log"))
	defer span.End()

	monitoring.Logger().DebugContext(
		ctx, "input", slog.Any("log", log),
	)

	realActorID, err := func() (*int64, error) {
		if log.ActorID == nil {
			return nil, nil
		}
		realID, err := strconv.ParseInt(*log.ActorID, 10, 64)
		if err != nil {
			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to parse actorID"),
			)
			return nil, err
		}
		return &realID, nil
	}()
	if err != nil {
		return err
	}

	query := `
		INSERT INTO
			immigration_clearance_logs (
				user_id,
				raw_response,
				status_code,
				first_name,
				last_name,
				gender,
				place_of_birth,
				date_of_birth,
				passport_number,
				country_code,
				is_in_indonesia,
				issued_at,
				expired_at,
				imei
			)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	args := []any{
		realActorID,
		string(log.RawResponse),
		log.StatusCode,
		log.FirstName,
		log.LastName,
		log.Gender,
		log.PlaceOfBirth,
		log.DateOfBirth,
		log.PassportNumber,
		log.CountryCode,
		log.IsInIndonesia,
		log.IssuedAt,
		log.ExpiredAt,
		log.IMEI.String(),
	}

	monitoring.Logger().DebugContext(
		ctx, "query", slog.String("query", query),
		"args", slog.Any("args", args),
	)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to insert log"),
		)
		return err
	}

	return nil
}

func (repo *ImmigrationClearanceLog) sName(s string) string {
	return "adapter.pg.immigration-clearance-log." + s
}
