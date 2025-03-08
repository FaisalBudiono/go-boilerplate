package portout

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"context"
)

type AuthActivityRepo interface {
	DeleteByPayload(ctx context.Context, tx domain.DBTX, payload string) error
	LastActivityByPayload(ctx context.Context, tx domain.DBTX, payload string) (string, error)
	Save(ctx context.Context, tx domain.DBTX, payload string, u domain.User) error
}
