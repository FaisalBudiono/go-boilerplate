package product

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/db"
	"context"
	"database/sql"
	"errors"

	"go.opentelemetry.io/otel/trace"
)

type (
	productFetcher interface {
		GetAll(
			ctx context.Context,
			tx db.DBTX,
			showAll bool,
			offset int64,
			limit int64,
		) ([]domain.Product, int64, error)
	}
	productIDFinder interface {
		FindByID(ctx context.Context, tx db.DBTX, id string) (domain.Product, error)
	}
	productSaver interface {
		Save(ctx context.Context, tx db.DBTX, name string, price int64) (domain.Product, error)
	}
	productPublisher interface {
		Publish(ctx context.Context, tx db.DBTX, p domain.Product, shouldPublish bool) (domain.Product, error)
	}
)

type Product struct {
	db     *sql.DB
	tracer trace.Tracer

	productFetcher   productFetcher
	productIDFinder  productIDFinder
	productPublisher productPublisher
	productSaver     productSaver
}

var (
	ErrEmptyName           = errors.New("Name cannot be empty")
	ErrNegativePrice       = errors.New("Price cannot be negative")
	ErrNotEnoughPermission = errors.New("Not enough permission")
	ErrNotFound            = errors.New("Product not found")
)

func New(
	db *sql.DB,
	tracer trace.Tracer,
	productFetcher productFetcher,
	productIDFinder productIDFinder,
	productPublisher productPublisher,
	productSaver productSaver,
) *Product {
	return &Product{
		db:     db,
		tracer: tracer,

		productFetcher:   productFetcher,
		productIDFinder:  productIDFinder,
		productPublisher: productPublisher,
		productSaver:     productSaver,
	}
}
