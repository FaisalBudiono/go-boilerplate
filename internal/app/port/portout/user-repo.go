package portout

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"context"
)

type UserRepo interface {
	FindByEmail(ctx context.Context, tx DBTX, email string) (domain.User, error)
	FindByID(ctx context.Context, tx DBTX, id domid.UserID) (domain.User, error)
}
