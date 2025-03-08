package res

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"math"
)

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

func lastPage(pg domain.Pagination) int64 {
	return int64(math.Ceil(float64(pg.Total) / float64(pg.PerPage)))
}
