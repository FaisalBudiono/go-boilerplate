package pg

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel/spanattr"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout"
	"context"
	"time"

	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/attribute"
)

type authActivity struct{}

func (repo *authActivity) DeleteByPayload(ctx context.Context, tx portout.DBTX, payload string) error {
	ctx, span := monitorings.Tracer().Start(ctx, "postgres: soft delete auth activity by payload")
	defer span.End()

	span.SetAttributes(attribute.String("input.payload", payload))

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

func (repo *authActivity) LastActivityByPayload(ctx context.Context, tx portout.DBTX, payload string) (domid.UserID, error) {
	ctx, span := monitorings.Tracer().Start(ctx, "postgres: update last activity by payload")
	defer span.End()

	span.SetAttributes(attribute.String("input.payload", payload))

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

func (repo *authActivity) Save(ctx context.Context, tx portout.DBTX, payload string, u domain.User) error {
	ctx, span := monitorings.Tracer().Start(ctx, "postgres: save auth_activities token")
	defer span.End()

	span.SetAttributes(attribute.String("input.payload", payload))
	span.SetAttributes(spanattr.User("input.", u)...)

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

func NewAuthActivity() *authActivity {
	return &authActivity{}
}
