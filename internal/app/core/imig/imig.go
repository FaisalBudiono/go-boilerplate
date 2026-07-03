package imig

import (
	"database/sql"
	"errors"

	"komdigi-immigration/internal/app/core/imig/eagerload"
	"komdigi-immigration/internal/app/core/logger/activitylog"
	"komdigi-immigration/internal/app/port"
)

func New(
	db *sql.DB,
	icLogRepo port.ImigrationClearanceLogRepo,
	icService port.ImmigrationClearanceService,
	activityLogRepo port.ActivityLogRepo,
	roleRepo port.RoleRepo,
	userRepo port.UserRepo,
) *Imig {
	return &Imig{
		db: db,

		icLogRepo: icLogRepo,
		icService: icService,

		activityLogger: activitylog.New(db, activityLogRepo),

		eagerLoader: eagerload.New(
			db,
			roleRepo,
			userRepo,
		),
	}
}

type Imig struct {
	db *sql.DB

	icLogRepo port.ImigrationClearanceLogRepo
	icService port.ImmigrationClearanceService

	activityLogger *activitylog.ActivityLog

	eagerLoader *eagerload.EagerLoad
}

func (srv *Imig) sName(s string) string {
	return "core.imig." + s
}

var (
	ErrPermissionDenied = errors.New("permission denied")

	ErrCountryCodeInvalid = errors.New("country code invalid")

	ErrLastNameMismatch    = errors.New("last name mismatch")
	ErrDateOfBirthMismatch = errors.New("date of birth mismatch")

	ErrClearanceInfoNotFound = errors.New("clearance information not found")
)
