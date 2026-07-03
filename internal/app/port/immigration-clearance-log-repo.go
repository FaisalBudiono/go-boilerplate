package port

import (
	"context"

	"komdigi-immigration/internal/app/domain"
)

type ImigrationClearanceLogRepo interface {
	GetIMEIHistoryPaginated(
		ctx context.Context, tx DBTX,
		imei string, page int64, perPage int64,
	) ([]domain.ImmigrationClearanceLog, int64, error)

	Log(
		ctx context.Context, tx DBTX, log domain.ImmigrationClearanceLogData,
	) error
}
