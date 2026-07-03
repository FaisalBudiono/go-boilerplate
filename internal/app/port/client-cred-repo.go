package port

import (
	"context"

	"komdigi-immigration/internal/app/domain"
	getallopt "komdigi-immigration/internal/app/port/options/clientcred/getall"
)

type ClientCredRepo interface {
	DeleteByClientID(ctx context.Context, tx DBTX, clientID string) error

	// FindByClientID can return [ErrDataNotFound]
	FindSecretByClientID(
		ctx context.Context, tx DBTX, clientID string,
	) (domain.ClientCredentialSecret, error)

	GetMapByClientID(
		ctx context.Context, tx DBTX, clientIDs []string,
	) (map[string]domain.ClientCredential, error)

	GetPaginated(
		ctx context.Context, tx DBTX,
		page, perPage int64,
		opts ...getallopt.QueryOption,
	) ([]domain.ClientCredential, int64, error)

	Insert(
		ctx context.Context, tx DBTX, data domain.ClientCredentialData,
	) error
}
