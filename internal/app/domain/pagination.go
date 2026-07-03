package domain

type Pagination struct {
	Page    int64
	PerPage int64
	Total   int64
}

func NewPagination(page, perPage, total int64) Pagination {
	return Pagination{
		Page:    page,
		PerPage: perPage,
		Total:   total,
	}
}
