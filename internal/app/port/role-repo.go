package port

import (
	"context"

	"FaisalBudiono/go-boilerplate/internal/app/domain"
)

type RoleRepo interface {
	AddByUserID(
		ctx context.Context, tx DBTX,
		userID string, roles []domain.Role,
	) error

	CleanByUserID(ctx context.Context, tx DBTX, userID string) error

	// GetMapByUserID returns a map of userID to slice of [domain.Role]
	GetMapByUserID(
		ctx context.Context, tx DBTX, userIDs []string,
	) (map[string][]domain.Role, error)
}
