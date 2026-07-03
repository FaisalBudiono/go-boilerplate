package port

import (
	"context"

	"komdigi-immigration/internal/app/domain"
)

type TokenRepo interface {
	Insert(ctx context.Context, tx DBTX, data domain.TokenCredentialData) error
	DeleteByClientID(ctx context.Context, tx DBTX, clientID string) error

	// FindByClientID might return an [ErrDataNotFound]
	FindByClientID(
		ctx context.Context, tx DBTX, clientID string,
	) (domain.TokenCredential, error)
}
