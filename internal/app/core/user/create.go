package user

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/mon"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/port"
)

type reqCreate interface {
	Context() context.Context
	Actor() domain.Userinfo

	Name() string
	Email() string
	Password() string
	Roles() []domain.Role
}

func (srv *User) Create(req reqCreate) (domain.UserEagerLoad, error) {
	ctx, span := mon.Tracer().Start(req.Context(), srv.sName("create"))
	defer span.End()

	actor := req.Actor()
	name := req.Name()
	email := req.Email()
	password := req.Password()
	roles := req.Roles()

	mon.Logger().InfoContext(
		ctx, "input",
		slog.Any("actor", actor),
		slog.String("name", name),
		slog.String("email", email),
		slog.Any("roles", roles),
	)

	emptyVal := domain.UserEagerLoad{}
	if !actor.User.HasRoles(domain.RoleAdmin) {
		return emptyVal, ErrPermissionDenied
	}

	tx, err := srv.db.BeginTx(ctx, nil)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to begin transaction"),
		)
		return emptyVal, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if !errors.Is(err, port.ErrTxDone) {
				otelutil.SpanLogError(
					span, err, otelutil.WithErrorLog(ctx),
					otelutil.WithMessage("failed to rollback transaction"),
				)
			}
		}
	}()

	_, err = srv.userRepo.FindByEmail(ctx, tx, email)
	if err == nil { // err IS NIL
		return emptyVal, ErrEmailDuplicated
	}
	if !errors.Is(err, sql.ErrNoRows) {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to find user by email"),
		)
		return emptyVal, err
	}

	hashed, err := srv.hasher.Hash(password)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to hash password"),
		)
		return emptyVal, err
	}

	userID, err := srv.userRepo.Insert(
		ctx, tx, domain.NewUserData(name, email, hashed),
	)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to insert user"),
		)
		return emptyVal, err
	}

	err = srv.roleRepo.AddByUserID(ctx, tx, userID, roles)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to add roles to user"),
		)
		return emptyVal, err
	}

	err = tx.Commit()
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to commit transaction"),
		)
		return emptyVal, err
	}

	userMap, err := srv.userRepo.GetMap(ctx, srv.db, []string{userID})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get user by id"),
		)
		return emptyVal, err
	}
	u := userMap[userID]

	eagerUsers, err := srv.eagerLoad.All(ctx, []domain.User{u})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to eager load users"),
		)
		return emptyVal, err
	}

	return eagerUsers[0], nil
}
