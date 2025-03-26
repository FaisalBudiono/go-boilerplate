package portout

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"context"
)

type AuthActivityRepo interface {
	DeleteByPayload(ctx context.Context, tx DBTX, payload string) error
	LastActivityByPayload(ctx context.Context, tx DBTX, payload string) (domid.UserID, error)
	Save(ctx context.Context, tx DBTX, payload string, u domain.User) error
}
