package pg

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/logutil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/codes"
)

type AuthActivity struct{}

func (repo *AuthActivity) DeleteByPayload(
	ctx context.Context, tx portout.DBTX, payload string,
) error {
	ctx, span := monitorings.Tracer().Start(ctx, "db.pg.AuthActivity.DeleteByPayload")
	defer span.End()

	monitorings.Logger().InfoContext(ctx, "input",
		slog.String("payload", payload),
	)

	now := time.Now().UTC()

	var userID string
	err := tx.QueryRowContext(
		ctx,
		`
UPDATE auth_activities
SET
    deleted_at = $1,
    updated_at = $2
WHERE
    deleted_at IS NULL
    AND payload = $3
RETURNING user_id
        `,
		now,
		now,
		payload,
	).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tracerr.CustomError(portout.ErrDataNotFound, tracerr.StackTrace(err))
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to query row")

		monitorings.Logger().ErrorContext(ctx, "failed to query row",
			slog.Any("error", err),
		)

		return tracerr.Wrap(err)
	}

	return nil
}

func (repo *AuthActivity) LastActivityByPayload(
	ctx context.Context, tx portout.DBTX, payload string,
) (domid.UserID, error) {
	ctx, span := monitorings.Tracer().Start(ctx, "db.pg.AuthActivity.LastActivityByPayload")
	defer span.End()

	monitorings.Logger().InfoContext(ctx, "input",
		slog.String("payload", payload),
	)

	now := time.Now().UTC()

	var userID string
	err := tx.QueryRowContext(
		ctx,
		`
UPDATE auth_activities
SET
    last_activity_at = $1,
    updated_at = $2
WHERE
    deleted_at IS NULL
    AND payload = $3
RETURNING user_id
        `,
		now,
		now,
		payload,
	).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", tracerr.CustomError(portout.ErrDataNotFound, tracerr.StackTrace(err))
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update row")

		monitorings.Logger().ErrorContext(ctx, "failed to update row",
			slog.Any("error", err),
		)

		return "", tracerr.Wrap(err)
	}

	return domid.UserID(userID), nil
}

func (repo *AuthActivity) Save(
	ctx context.Context, tx portout.DBTX, payload string, u domain.User,
) error {
	ctx, span := monitorings.Tracer().Start(ctx, "db.pg.AuthActivity.Save")
	defer span.End()

	logVals := []any{slog.String("payload", payload)}
	logVals = append(logVals, logutil.SlogUser(u)...)

	monitorings.Logger().InfoContext(ctx, "input",
		logVals...,
	)

	_, err := tx.ExecContext(
		ctx,
		`
INSERT INTO
    auth_activities (user_id, payload, last_activity_at)
VALUES
    ($1, $2, $3)
        `,
		u.ID, payload, time.Now().UTC(),
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to inserting auth activity to db")

		monitorings.Logger().ErrorContext(ctx, "failed to inserting auth activity to db",
			slog.Any("error", err),
		)

		return tracerr.Wrap(err)
	}

	return nil
}

func NewAuthActivity() *AuthActivity {
	return &AuthActivity{}
}
