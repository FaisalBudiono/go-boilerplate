package user

import (
	"database/sql"
	"errors"

	"komdigi-immigration/internal/app/core/user/eagerload"
	"komdigi-immigration/internal/app/port"
)

func New(
	db *sql.DB,
	userRepo port.UserRepo,
	roleRepo port.RoleRepo,
) *User {
	return &User{
		db:        db,
		userRepo:  userRepo,
		eagerLoad: eagerload.New(db, roleRepo),
	}
}

type User struct {
	db *sql.DB

	userRepo port.UserRepo

	eagerLoad *eagerload.EagerLoad
}

const (
	defaultPage    int64 = 1
	defaultPerPage int64 = 10
)

var ErrPermissionDenied = errors.New("permission denied")
