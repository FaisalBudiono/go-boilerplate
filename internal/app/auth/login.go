package auth

import (
	"FaisalBudiono/go-boilerplate/internal/app/rnd"
	"FaisalBudiono/go-boilerplate/internal/domain"
	"context"
	"database/sql"
	"errors"

	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/attribute"
)

type inputLogin interface {
	Context() context.Context
	Email() string
	Password() string
}

func (srv *Auth) Login(req inputLogin) (domain.Token, error) {
	ctx, span := srv.tracer.Start(req.Context(), "service: login")
	defer span.End()

	email := req.Email()
	span.SetAttributes(attribute.String("input.email", email))

	tx, err := srv.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Token{}, tracerr.Wrap(err)
	}
	defer tx.Rollback()

	u, err := srv.userEmailFinder.FindByEmail(ctx, tx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Token{}, tracerr.CustomError(
				ErrInvalidCredentials,
				tracerr.StackTrace(err),
			)
		}

		return domain.Token{}, err
	}

	match, err := srv.passwordVerifier.Verify(req.Password(), u.Password)
	if err != nil {
		return domain.Token{}, err
	}

	if !match {
		return domain.Token{}, tracerr.Wrap(ErrInvalidCredentials)
	}

	token, err := srv.jwtUserSigner.Sign(domain.NewUserBasicInfo(u.ID))
	if err != nil {
		return domain.Token{}, err
	}

	refreshTokenPayload := rnd.UUID()
	err = srv.authActivitySaver.Save(ctx, tx, refreshTokenPayload, u)
	if err != nil {
		return domain.Token{}, err
	}

	refreshToken, err := srv.refreshTokenSigner.Sign(refreshTokenPayload)
	if err != nil {
		return domain.Token{}, err
	}

	err = tx.Commit()
	if err != nil {
		return domain.Token{}, tracerr.Wrap(err)
	}

	return domain.NewToken(token, refreshToken), nil
}
