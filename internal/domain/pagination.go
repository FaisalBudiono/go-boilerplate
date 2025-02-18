package domain

type Pagination struct {
	Page    int64
	PerPage int64
	Total   int64
}

func NewPagination(
	Page int64,
	PerPage int64,
	Total int64,
) Pagination {
	return Pagination{
		Page:    Page,
		PerPage: PerPage,
		Total:   Total,
	}
}
