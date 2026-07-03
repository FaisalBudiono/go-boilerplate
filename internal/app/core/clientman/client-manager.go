package clientman

import (
	"database/sql"
	"errors"
	"fmt"

	"komdigi-immigration/internal/app/core/auth/passwd"
	"komdigi-immigration/internal/app/core/clientman/eagerload"
	"komdigi-immigration/internal/app/core/logger/activitylog"
	"komdigi-immigration/internal/app/port"
)

func New(
	db *sql.DB,
	userRepo port.UserRepo,
	roleRepo port.RoleRepo,
	clientCredRepo port.ClientCredRepo,
	activityLogRepo port.ActivityLogRepo,
	hasher passwd.Hasher,
) *ClientManager {
	return &ClientManager{
		db: db,

		userRepo:       userRepo,
		roleRepo:       roleRepo,
		clientCredRepo: clientCredRepo,

		hasher: hasher,

		eagerLoad:      eagerload.New(db, roleRepo, userRepo),
		activityLogger: activitylog.New(db, activityLogRepo),
	}
}

type ClientManager struct {
	db *sql.DB

	userRepo       port.UserRepo
	roleRepo       port.RoleRepo
	clientCredRepo port.ClientCredRepo

	hasher passwd.Hasher

	eagerLoad      *eagerload.EagerLoad
	activityLogger *activitylog.ActivityLog
}

func (srv *ClientManager) spanName(s string) string {
	return fmt.Sprintf("core.clientman.%s", s)
}

var (
	ErrPermissionDenied = errors.New("permission denied")

	ErrCannotGrantAdmin = errors.New("cannot grant access for admin")

	ErrUserNotFound = errors.New("user not found")

	ErrClientIDNotFound = errors.New("clientID not found")
)
