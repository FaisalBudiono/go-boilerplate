package port

import (
	"context"

	"komdigi-immigration/internal/app/domain"
)

type RoleRepo interface {
	// GetMapByUserID returns a map of userID to slice of [domain.Role]
	GetMapByUserID(
		ctx context.Context, tx DBTX, userIDs []string,
	) (map[string][]domain.Role, error)
}
