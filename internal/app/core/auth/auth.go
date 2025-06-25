package auth

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout"
	"database/sql"
	"errors"
)

var (
	ErrInvalidCredentials = errors.New("Invalid credentials")
	ErrInvalidToken       = errors.New("Invalid token")
	ErrTokenExpired       = errors.New("Token expired")
)

type (
	passwordVerifier interface {
		Verify(password, encodedHash string) (match bool, err error)
	}

	jwtUserSigner interface {
		Sign(u domain.UserTokenInfo) (string, error)
	}

	jwtUserParser interface {
		Parse(token string) (domain.UserTokenInfo, error)
	}

	refreshTokenSigner interface {
		Sign(payload string) (string, error)
	}

	refreshTokenPayloadParser interface {
		ParsePayload(token string) (string, error)
	}
)

type Auth struct {
	db *sql.DB

	authActivityRepo portout.AuthActivityRepo
	userRepo         portout.UserRepo

	passwordVerifier passwordVerifier

	jwtUserSigner jwtUserSigner
	jwtUserParser jwtUserParser

	refreshTokenSigner        refreshTokenSigner
	refreshTokenPayloadParser refreshTokenPayloadParser
}

func New(
	db *sql.DB,
	authActivityRepo portout.AuthActivityRepo,
	userRepo portout.UserRepo,
	passwordVerifier passwordVerifier,
	jwtUserSigner jwtUserSigner,
	jwtUserParser jwtUserParser,
	refreshTokenSigner refreshTokenSigner,
	refreshTokenPayloadParser refreshTokenPayloadParser,
) *Auth {
	return &Auth{
		db: db,

		authActivityRepo: authActivityRepo,
		userRepo:         userRepo,

		passwordVerifier: passwordVerifier,

		jwtUserSigner: jwtUserSigner,
		jwtUserParser: jwtUserParser,

		refreshTokenSigner:        refreshTokenSigner,
		refreshTokenPayloadParser: refreshTokenPayloadParser,
	}
}
