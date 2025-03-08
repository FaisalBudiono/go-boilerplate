package res

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"time"
)

type product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       int64   `json:"price"`
	PublishedAt *string `json:"publishedAt"`
}

func ToProduct(p domain.Product) response[product] {
	var pubAt *string
	if p.PublishedAt != nil {
		f := p.PublishedAt.Format(time.RFC3339Nano)
		pubAt = &f
	}

	return response[product]{
		Data: product{
			ID:          string(p.ID),
			Name:        p.Name,
			Price:       p.Price / 100,
			PublishedAt: pubAt,
		},
	}
}

func ToProductPaginated(ps []domain.Product, pg domain.Pagination) responsePaginated[product] {
	resProducts := make([]product, len(ps))
	for i, p := range ps {
		resProducts[i] = toProductRes(p)
	}

	return responsePaginated[product]{
		response: response[[]product]{
			Data: resProducts,
		},
		Meta: paginatedMeta{
			Page:     pg.Page,
			Total:    pg.Total,
			PerPage:  pg.PerPage,
			LastPage: lastPage(pg),
		},
	}
}

func toProductRes(p domain.Product) product {
	var pubAt *string
	if p.PublishedAt != nil {
		f := p.PublishedAt.Format(time.RFC3339Nano)
		pubAt = &f
	}

	return product{
		ID:          string(p.ID),
		Name:        p.Name,
		Price:       p.Price / 100,
		PublishedAt: pubAt,
	}
}
