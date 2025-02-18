package domain

import "time"

type Product struct {
	ID          string
	Name        string
	Price       int64
	PublishedAt *time.Time
}

func NewProduct(
	id string,
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
