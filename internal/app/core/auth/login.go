package auth

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/rnd"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout"
	"context"
	"errors"
	"log/slog"
)

type inputLogin interface {
	Context() context.Context
	Email() string
	Password() string
}

func (srv *Auth) Login(req inputLogin) (domain.Token, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), "core.Auth.Login")
	defer span.End()

	email := req.Email()

	monitoring.Logger().InfoContext(ctx, "input", slog.String("email", email))

	tx, err := srv.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Token{}, err
	}
	defer tx.Rollback()

	u, err := srv.userRepo.FindByEmail(ctx, tx, email)
	if err != nil {
		if errors.Is(err, portout.ErrDataNotFound) {
			return domain.Token{}, errors.Join(ErrInvalidCredentials, err)
		}

		return domain.Token{}, err
	}

	match, err := srv.passwordVerifier.Verify(req.Password(), u.Password)
	if err != nil {
		return domain.Token{}, err
	}

	if !match {
		return domain.Token{}, ErrInvalidCredentials
	}

	token, err := srv.jwtUserSigner.Sign(domain.NewUserBasicInfo(u.ID))
	if err != nil {
		return domain.Token{}, err
	}

	refreshTokenPayload := rnd.UUID()
	err = srv.authActivityRepo.Save(ctx, tx, refreshTokenPayload, u)
	if err != nil {
		return domain.Token{}, err
	}

	refreshToken, err := srv.refreshTokenSigner.Sign(refreshTokenPayload)
	if err != nil {
		return domain.Token{}, err
	}

	err = tx.Commit()
	if err != nil {
		return domain.Token{}, err
	}

	return domain.NewToken(token, refreshToken), nil
}
