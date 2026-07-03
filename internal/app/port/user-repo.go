package port

import (
	"context"

	"komdigi-immigration/internal/app/domain"
	getallopt "komdigi-immigration/internal/app/port/options/user/getall"
)

type UserRepo interface {
	FindByEmail(ctx context.Context, tx DBTX, email string) (domain.UserWithPassword, error)

	// GetMap returns a map of userID to [domain.User]
	GetMap(ctx context.Context, tx DBTX, ids []string) (map[string]domain.User, error)

	// GetPaginated returns a paginated list of users with total
	GetPaginated(
		ctx context.Context, tx DBTX,
		page, perPage int64,
		opts ...getallopt.QueryOption,
	) ([]domain.User, int64, error)
}
