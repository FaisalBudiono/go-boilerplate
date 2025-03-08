package product

import (
	"FaisalBudiono/go-boilerplate/internal/app/port/portout"
	"database/sql"
	"errors"

	"go.opentelemetry.io/otel/trace"
)

type Product struct {
	db     *sql.DB
	tracer trace.Tracer

	productRepo portout.ProductRepo
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
	productRepo portout.ProductRepo,
) *Product {
	return &Product{
		db:     db,
		tracer: tracer,

		productRepo: productRepo,
	}
}
