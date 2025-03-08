package portout

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"context"
)

type UserRepo interface {
	FindByEmail(ctx context.Context, tx domain.DBTX, email string) (domain.User, error)
	FindByID(ctx context.Context, tx domain.DBTX, id string) (domain.User, error)
}
