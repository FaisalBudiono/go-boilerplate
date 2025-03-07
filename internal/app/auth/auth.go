package auth

import (
	"FaisalBudiono/go-boilerplate/internal/db"
	"FaisalBudiono/go-boilerplate/internal/domain"
	"context"
	"database/sql"
	"errors"

	"go.opentelemetry.io/otel/trace"
)

var (
	ErrInvalidCredentials = errors.New("Invalid credentials")
	ErrInvalidToken       = errors.New("Invalid token")
	ErrTokenExpired       = errors.New("Token expired")
)

type (
	userIDFinder interface {
		FindByID(ctx context.Context, tx db.DBTX, id string) (domain.User, error)
	}

	userEmailFinder interface {
		FindByEmail(ctx context.Context, tx db.DBTX, email string) (domain.User, error)
	}

	authActivitySaver interface {
		Save(ctx context.Context, tx db.DBTX, payload string, u domain.User) error
	}

	authActivityLastActivityUpdater interface {
		LastActivityByPayload(ctx context.Context, tx db.DBTX, payload string) (string, error)
	}

	authActivityPayloadDeleter interface {
		DeleteByPayload(ctx context.Context, tx db.DBTX, payload string) error
	}

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
	db    *sql.DB
	tracer trace.Tracer

	authActivitySaver               authActivitySaver
	authActivityLastActivityUpdater authActivityLastActivityUpdater
	authActivityPayloadDeleter      authActivityPayloadDeleter

	userIDFinder    userIDFinder
	userEmailFinder userEmailFinder

	passwordVerifier passwordVerifier

	jwtUserSigner jwtUserSigner
	jwtUserParser jwtUserParser

	refreshTokenSigner        refreshTokenSigner
	refreshTokenPayloadParser refreshTokenPayloadParser
}

func New(
	db *sql.DB,
	tracer trace.Tracer,
	authActivitySaver authActivitySaver,
	authActivityLastActivityUpdater authActivityLastActivityUpdater,
	authActivityPayloadDeleter authActivityPayloadDeleter,
	userIDFinder userIDFinder,
	userEmailFinder userEmailFinder,
	passwordVerifier passwordVerifier,
	jwtUserSigner jwtUserSigner,
	jwtUserParser jwtUserParser,
	refreshTokenSigner refreshTokenSigner,
	refreshTokenPayloadParser refreshTokenPayloadParser,
) *Auth {
	return &Auth{
		db:    db,
		tracer: tracer,

		authActivitySaver:               authActivitySaver,
		authActivityLastActivityUpdater: authActivityLastActivityUpdater,
		authActivityPayloadDeleter:      authActivityPayloadDeleter,

		userIDFinder:    userIDFinder,
		userEmailFinder: userEmailFinder,

		passwordVerifier: passwordVerifier,

		jwtUserSigner: jwtUserSigner,
		jwtUserParser: jwtUserParser,

		refreshTokenSigner:        refreshTokenSigner,
		refreshTokenPayloadParser: refreshTokenPayloadParser,
	}
}
