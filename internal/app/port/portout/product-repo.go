package portout

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout/options/product/optproduct"
	"context"
)

type ProductRepo interface {
	// FindByID will find [domain.Product] by its ID
	FindByID(ctx context.Context, tx DBTX, id string) (domain.Product, error)
	GetAll(ctx context.Context, tx DBTX, offset, limit int64, qo ...optproduct.QueryOption) ([]domain.Product, int64, error)
	Publish(ctx context.Context, tx DBTX, p domain.Product, shouldPublish bool) (domain.Product, error)
	Save(ctx context.Context, tx DBTX, name string, price int64) (domain.Product, error)
}
