package eagerload

import (
	"database/sql"
	"fmt"

	"komdigi-immigration/internal/app/port"
)

func New(
	db *sql.DB,
	roleRepo port.RoleRepo,
	userRepo port.UserRepo,
) *EagerLoad {
	return &EagerLoad{
		db:       db,
		roleRepo: roleRepo,
		userRepo: userRepo,
	}
}

type EagerLoad struct {
	db *sql.DB

	roleRepo port.RoleRepo
	userRepo port.UserRepo
}

func (el *EagerLoad) spanName(s string) string {
	return fmt.Sprintf("core.clientman.eagerload.%s", s)
}
