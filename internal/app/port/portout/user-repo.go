package portout

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"context"
)

type UserRepo interface {
	FindByEmail(ctx context.Context, tx domain.DBTX, email string) (domain.User, error)
	FindByID(ctx context.Context, tx domain.DBTX, id domid.UserID) (domain.User, error)
}
