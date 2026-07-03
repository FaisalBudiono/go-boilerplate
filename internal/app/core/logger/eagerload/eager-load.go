package eagerload

import (
	"database/sql"

	"komdigi-immigration/internal/app/port"
)

func New(
	db *sql.DB,
	actLogRepo port.ActivityLogRepo,
	userRepo port.UserRepo,
	roleRepo port.RoleRepo,
) *EagerLoad {
	return &EagerLoad{
		db:         db,
		actLogRepo: actLogRepo,
		userRepo:   userRepo,
		roleRepo:   roleRepo,
	}
}

type EagerLoad struct {
	db *sql.DB

	actLogRepo port.ActivityLogRepo
	userRepo   port.UserRepo
	roleRepo   port.RoleRepo
}

func (el *EagerLoad) sName(s string) string {
	return "core.logger.eagerload." + s
}
