package portout

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout/productoptions"
	"context"
)

type ProductRepo interface {
	FindByID(ctx context.Context, tx DBTX, id domid.ProductID) (domain.Product, error)
	GetAll(ctx context.Context, tx DBTX, offset, limit int64, qo ...productoptions.QueryOption) ([]domain.Product, int64, error)
	Publish(ctx context.Context, tx DBTX, p domain.Product, shouldPublish bool) (domain.Product, error)
	Save(ctx context.Context, tx DBTX, name string, price int64) (domain.Product, error)
}
