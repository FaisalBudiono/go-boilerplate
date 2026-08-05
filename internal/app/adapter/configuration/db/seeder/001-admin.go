package seeder

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"FaisalBudiono/go-boilerplate/internal/app/core/auth/passwd"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/mon"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/queryutil"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/port"
)

func NewAdmin(
	db *sql.DB,
	adminName string,
	adminEmail string,
	adminPassword string,
	hasher passwd.Hasher,
) *admin {
	return &admin{
		db: db,

		adminName:     adminName,
		adminEmail:    adminEmail,
		adminPassword: adminPassword,

		hasher: hasher,
	}
}

type admin struct {
	db *sql.DB

	adminName     string
	adminEmail    string
	adminPassword string

	hasher passwd.Hasher
}

func (srv *admin) Name() string { return "admin" }

func (srv *admin) Seed(ctx context.Context) error {
	ctx, span := mon.Tracer().Start(ctx, "seeder.admin.seed")
	defer span.End()

	hashedPassword, err := srv.hasher.Hash(srv.adminPassword)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to hash password"),
		)
		return err
	}

	hasAdmin, err := srv.hasAdmin(ctx)
	if err != nil {
		return err
	}

	upsert := func() adminUpsertFunc {
		if hasAdmin {
			return srv.update
		}
		return srv.insert
	}()

	tx, err := srv.db.BeginTx(ctx, nil)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to begin transaction"),
		)
		return err
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

	err = upsert(ctx, tx, hashedPassword)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to upsert admin"),
		)
		return err
	}

	err = srv.deleteRole(ctx, tx)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to delete role"),
		)
		return err
	}

	err = srv.insertRole(ctx, tx)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to insert role"),
		)
		return err
	}

	err = tx.Commit()
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to commit transaction"),
		)
		return err
	}

	return nil
}

func (srv *admin) deleteRole(ctx context.Context, tx port.DBTX) error {
	ctx, span := mon.Tracer().Start(ctx, "seeder.admin.deleteRole")
	defer span.End()

	query := `
		DELETE FROM
			user_roles
		WHERE
			user_id = $1
	`
	args := []any{idAdmin}

	mon.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)
	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to delete role"),
		)
		return err
	}

	return nil
}

func (srv *admin) insertRole(ctx context.Context, tx port.DBTX) error {
	ctx, span := mon.Tracer().Start(ctx, "seeder.admin.insertRole")
	defer span.End()

	query := `
		INSERT INTO
			user_roles (user_id, name)
		VALUES
			($1, $2)
	`
	args := []any{idAdmin, domain.RoleAdmin.String()}

	mon.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to insert role"),
		)
		return err
	}

	return nil
}

type adminUpsertFunc func(ctx context.Context, tx port.DBTX, hashedPassword string) error

func (srv *admin) insert(ctx context.Context, tx port.DBTX, hashedPassword string) error {
	ctx, span := mon.Tracer().Start(ctx, "seeder.admin.insert")
	defer span.End()

	query := `
		INSERT INTO
			users (id, name, email, password)
		VALUES
			($1, $2, $3, $4)
	`
	args := []any{idAdmin, srv.adminName, srv.adminEmail, hashedPassword}

	mon.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to insert admin"),
		)
		return err
	}

	return nil
}

func (srv *admin) update(ctx context.Context, tx port.DBTX, hashedPassword string) error {
	ctx, span := mon.Tracer().Start(ctx, "seeder.admin.update")
	defer span.End()

	query := `
		UPDATE
			users
		SET
			name = $1,
			email = $2,
			password = $3,
			deleted_at = NULL
		WHERE
			id = $4
	`
	args := []any{srv.adminName, srv.adminEmail, hashedPassword, idAdmin}

	mon.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to update admin"),
		)
		return err
	}

	return nil
}

func (srv *admin) hasAdmin(ctx context.Context) (bool, error) {
	ctx, span := mon.Tracer().Start(ctx, "seeder.admin.hasAdmin")
	defer span.End()

	var id int64
	err := srv.db.QueryRowContext(ctx, "SELECT id FROM users WHERE id = $1", idAdmin).Scan(&id)
	if err != nil {
		if err != sql.ErrNoRows {
			otelutil.SpanLogError(
				span, err,
				otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to check if admin exists"),
			)
			return false, err
		}

		return false, nil
	}

	return true, nil
}
