package logger

import (
	"database/sql"
	"errors"

	"komdigi-immigration/internal/app/core/logger/eagerload"
	"komdigi-immigration/internal/app/port"
)

func New(
	db *sql.DB,
	activityLogRepo port.ActivityLogRepo,
	userRepo port.UserRepo,
	roleRepo port.RoleRepo,
) *Logger {
	return &Logger{
		db:              db,
		activityLogRepo: activityLogRepo,

		eagerLoader: eagerload.New(
			db, activityLogRepo, userRepo, roleRepo,
		),
	}
}

type Logger struct {
	db *sql.DB

	activityLogRepo port.ActivityLogRepo

	eagerLoader *eagerload.EagerLoad
}

func (srv *Logger) sName(s string) string {
	return "core.logger." + s
}

const (
	defaultPage    int64 = 1
	defaultPerPage int64 = 10
)

var ErrPermissionDenied = errors.New("permission denied")
