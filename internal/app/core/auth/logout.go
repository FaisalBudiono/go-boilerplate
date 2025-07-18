package auth

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/auth/jwt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout"
	"context"
	"errors"

	"github.com/ztrue/tracerr"
)

type inputLogout interface {
	Context() context.Context
	RefreshToken() string
}

func (srv *Auth) Logout(req inputLogout) error {
	ctx, span := monitorings.Tracer().Start(req.Context(), "core.auth.logout")
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

	err = srv.authActivityRepo.DeleteByPayload(ctx, tx, payload)
	if err != nil {
		if errors.Is(err, portout.ErrDataNotFound) {
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
