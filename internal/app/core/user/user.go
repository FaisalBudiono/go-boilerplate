package user

import (
	"database/sql"
	"errors"

	"FaisalBudiono/go-boilerplate/internal/app/core/auth/passwd"
	"FaisalBudiono/go-boilerplate/internal/app/core/user/eagerload"
	"FaisalBudiono/go-boilerplate/internal/app/port"
)

func New(
	db *sql.DB,
	userRepo port.UserRepo,
	roleRepo port.RoleRepo,
	hasher passwd.Hasher,
) *User {
	return &User{
		db: db,

		userRepo: userRepo,
		roleRepo: roleRepo,

		hasher: hasher,

		eagerLoad: eagerload.New(db, roleRepo),
	}
}

type User struct {
	db *sql.DB

	userRepo port.UserRepo
	roleRepo port.RoleRepo

	hasher passwd.Hasher

	eagerLoad *eagerload.EagerLoad
}

func (srv *User) sName(s string) string {
	return "core.user." + s
}

const (
	defaultPage    int64 = 1
	defaultPerPage int64 = 10
)

var (
	ErrPermissionDenied = errors.New("permission denied")

	ErrEmailDuplicated = errors.New("email already exists")
)
