package port

import (
	"context"

	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/domain/actlog"
	"komdigi-immigration/internal/app/port/options/activitylog/getall"
)

type ActivityLogRepo interface {
	AddMetadata(
		ctx context.Context,
		tx DBTX,
		id string,
		metadata map[actlog.MetaKey]string,
	) error

	// GetPaginated returns a paginated list of activityLog with total
	GetPaginated(
		ctx context.Context, tx DBTX,
		page, perPage int64,
		opts ...getall.QueryOption,
	) ([]domain.ActivityLog, int64, error)

	GetKeyMapByLogIDs(
		ctx context.Context, tx DBTX, logIDs []string,
	) (map[string]map[actlog.MetaKey]string, error)

	// Insert return activity log id
	Insert(
		ctx context.Context, tx DBTX, data domain.ActivityLogData,
	) (string, error)
}
