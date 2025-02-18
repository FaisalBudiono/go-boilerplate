package auth

import (
	"FaisalBudiono/go-boilerplate/internal/app/auth/jwt"
	"context"
	"database/sql"
	"errors"

	"github.com/ztrue/tracerr"
)

type inputLogout interface {
	Context() context.Context
	RefreshToken() string
}

func (srv *Auth) Logout(req inputLogout) error {
	ctx, span := srv.tracer.Start(req.Context(), "service: logout")
	defer span.End()

	payload, err := srv.refreshTokenPayloadParser.ParsePayload(req.RefreshToken())
	if err != nil {
		isInvalidTokenErr := errors.Is(err, jwt.ErrTokenMalformed) ||
			errors.Is(err, jwt.ErrSignatureInvalid) ||
			errors.Is(err, jwt.ErrTokenExpired)

		if isInvalidTokenErr {
			return tracerr.CustomError(ErrInvalidToken, tracerr.StackTrace(err))
		}

		return err
	}

	tx, err := srv.db.BeginTx(ctx, nil)
	if err != nil {
		return tracerr.Wrap(err)
	}
	defer tx.Rollback()

	err = srv.authActivityPayloadDeleter.DeleteByPayload(ctx, tx, payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tracerr.CustomError(ErrTokenExpired, tracerr.StackTrace(err))
		}

		return err
	}

	err = tx.Commit()
	if err != nil {
		return tracerr.Wrap(err)
	}

	return nil
}
