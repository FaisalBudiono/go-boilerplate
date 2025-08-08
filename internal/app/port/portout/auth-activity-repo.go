package portout

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"context"
)

type AuthActivityRepo interface {
	DeleteByPayload(ctx context.Context, tx DBTX, payload string) error
	// LastActivityByPayload will return user ID
	LastActivityByPayload(ctx context.Context, tx DBTX, payload string) (string, error)
	Save(ctx context.Context, tx DBTX, payload string, u domain.User) error
}
