package pg

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/logutil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout"
	"context"
	"log/slog"
	"time"

	"github.com/ztrue/tracerr"
)

type AuthActivity struct{}

func (repo *AuthActivity) DeleteByPayload(ctx context.Context, tx portout.DBTX, payload string) error {
	ctx, span := monitorings.Tracer().Start(ctx, "db.pg.auth-activity.deleteByPayload")
	defer span.End()

	monitorings.Logger().InfoContext(
		ctx,
		"input",
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
		return tracerr.Wrap(err)
	}

	return nil
}

func (repo *AuthActivity) LastActivityByPayload(ctx context.Context, tx portout.DBTX, payload string) (domid.UserID, error) {
	ctx, span := monitorings.Tracer().Start(ctx, "db.pg.auth-activity.lastActivityByPayload")
	defer span.End()

	monitorings.Logger().InfoContext(
		ctx,
		"input",
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
		return "", tracerr.Wrap(err)
	}

	return domid.UserID(userID), nil
}

func (repo *AuthActivity) Save(ctx context.Context, tx portout.DBTX, payload string, u domain.User) error {
	ctx, span := monitorings.Tracer().Start(ctx, "db.pg.auth-activity.save")
	defer span.End()

	logVals := []any{slog.String("payload", payload)}
	logVals = append(logVals, logutil.SlogUser(u)...)

	monitorings.Logger().InfoContext(
		ctx,
		"input",
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

	return tracerr.Wrap(err)
}

func NewAuthActivity() *AuthActivity {
	return &AuthActivity{}
}
