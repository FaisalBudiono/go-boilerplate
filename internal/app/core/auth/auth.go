package auth

import (
	"database/sql"
	"errors"
	"fmt"

	"komdigi-immigration/internal/app/core/auth/passwd"
	"komdigi-immigration/internal/app/core/logger/activitylog"
	"komdigi-immigration/internal/app/core/user/eagerload"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/port"
)

func New(
	db *sql.DB,
	userRepo port.UserRepo,
	tokenRepo port.TokenRepo,
	roleRepo port.RoleRepo,
	activityLogRepo port.ActivityLogRepo,
	clientCredRepo port.ClientCredRepo,
	hasher passwd.Hasher,
	jwtUserSigner jwtUserSigner,
	jwtRefreshSigner jwtRefreshSigner,
) *Auth {
	return &Auth{
		db: db,

		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		roleRepo:       roleRepo,
		clientCredRepo: clientCredRepo,

		hasher:           hasher,
		jwtUserSigner:    jwtUserSigner,
		jwtRefreshSigner: jwtRefreshSigner,

		userLoader:     eagerload.New(db, roleRepo),
		activityLogger: activitylog.New(db, activityLogRepo),
	}
}

type Auth struct {
	db *sql.DB

	userRepo       port.UserRepo
	tokenRepo      port.TokenRepo
	roleRepo       port.RoleRepo
	clientCredRepo port.ClientCredRepo

	hasher           passwd.Hasher
	jwtUserSigner    jwtUserSigner
	jwtRefreshSigner jwtRefreshSigner

	userLoader     *eagerload.EagerLoad
	activityLogger *activitylog.ActivityLog
}

func (srv *Auth) spanName(s string) string {
	return fmt.Sprintf("core.auth.%s", s)
}

type jwtUserSigner interface {
	Sign(u domain.UserTokenInfo) (string, error)
	Parse(token string) (domain.UserTokenInfo, error)
}

type jwtRefreshSigner interface {
	Sign(id, secret string) (string, error)
	Parse(token string) (domain.RefreshToken, error)
}

var (
	ErrEmailRequired    = errors.New("email is required")
	ErrPasswordRequired = errors.New("password is required")

	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDeniedEmailLogin   = errors.New("login with email denied")

	ErrTokenExpired = errors.New("access token expired")
	ErrTokenInvalid = errors.New("access token invalid")

	ErrUserNotFound = errors.New("user not found")
)
