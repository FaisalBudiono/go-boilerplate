package portout

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"context"
)

type UserRepo interface {
	FindByEmail(ctx context.Context, tx DBTX, email string) (domain.User, error)
	// FindByID will find [domain.User] by its ID
	FindByID(ctx context.Context, tx DBTX, id string) (domain.User, error)
}
