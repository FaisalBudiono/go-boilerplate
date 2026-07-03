package eagerload

import (
	"FaisalBudiono/go-boilerplate/internal/app/port"
)

func New(
	db port.DBTX,
	roleRepo port.RoleRepo,
) *EagerLoad {
	return &EagerLoad{
		db:       db,
		roleRepo: roleRepo,
	}
}

type EagerLoad struct {
	db port.DBTX

	roleRepo port.RoleRepo
}

func (el *EagerLoad) sName(s string) string {
	return "core.user.eagerload." + s
}
