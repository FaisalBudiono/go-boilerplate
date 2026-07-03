package res

import (
	"math"

	"komdigi-immigration/internal/app/domain"
)

type response[T any] struct {
	Data T `json:"data"`
}

type paginatedMeta struct {
	Page     int64 `json:"page"`
	PerPage  int64 `json:"perPage"`
	LastPage int64 `json:"lastPage"`
	Total    int64 `json:"total"`
}

type responsePaginated[T any] struct {
	response[[]T]
	Meta paginatedMeta `json:"meta"`
}

func newPaginatedMeta(pg domain.Pagination) paginatedMeta {
	return paginatedMeta{
		Page:     pg.Page,
		Total:    pg.Total,
		PerPage:  pg.PerPage,
		LastPage: int64(math.Ceil(float64(pg.Total) / float64(pg.PerPage))),
	}
}
