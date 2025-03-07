package domain

import (
	"FaisalBudiono/go-boilerplate/internal/domain/domid"
	"time"
)

type Product struct {
	ID          domid.ProductID
	Name        string
	Price       int64
	PublishedAt *time.Time
}

func NewProduct(
	id domid.ProductID,
	name string,
	price int64,
	publishedAt *time.Time,
) Product {
	return Product{
		ID:          id,
		Name:        name,
		Price:       price,
		PublishedAt: publishedAt,
	}
}
