package portout

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"context"
)

type ProductRepo interface {
	FindByID(ctx context.Context, tx domain.DBTX, id domid.ProductID) (domain.Product, error)
	GetAll(ctx context.Context, tx domain.DBTX, showAll bool, offset int64, limit int64) ([]domain.Product, int64, error)
	Publish(ctx context.Context, tx domain.DBTX, p domain.Product, shouldPublish bool) (domain.Product, error)
	Save(ctx context.Context, tx domain.DBTX, name string, price int64) (domain.Product, error)
}
