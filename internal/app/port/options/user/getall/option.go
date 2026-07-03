package getall

import "komdigi-immigration/internal/app/domain"

type queryOpt struct {
	Roles []domain.Role
}

func NewQueryOpt() *queryOpt {
	return &queryOpt{
		Roles: []domain.Role{},
	}
}

type QueryOption func(*queryOpt)

func WithRoleFlags(roles ...domain.Role) QueryOption {
	return func(qo *queryOpt) {
		qo.Roles = roles
	}
}
