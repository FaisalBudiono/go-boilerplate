package pg

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout"
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/ztrue/tracerr"
)

type User struct {
	r *Role
}

type resultRoleMap struct {
	res map[domid.UserID][]domain.Role
	err error
}

// Not safe for transactions
func (repo *User) FindByID(ctx context.Context, tx portout.DBTX, id domid.UserID) (domain.User, error) {
	ctx, span := monitorings.Tracer().Start(ctx, "db.pg.User.findByID")
	defer span.End()

	monitorings.Logger().InfoContext(ctx, "input", slog.String("id", string(id)))

	ctx, cancel := context.WithCancelCause(ctx)

	chanRoleRes := make(chan resultRoleMap)
	go func() {
		rMap, err := repo.r.ByUserIDs(ctx, tx, []domid.UserID{id})
		if err != nil {
			cancel(err)
		}

		chanRoleRes <- resultRoleMap{rMap, err}
	}()

	q := `
SELECT
    id,
    name,
    email,
    phone_number,
    password
FROM
    users
WHERE
    id = $1
LIMIT
    1
`
	monitorings.Logger().DebugContext(ctx, "making query", slog.String("query", q))

	var raw struct {
		id          string
		name        string
		phoneNumber string
		email       string
		password    string
	}
	err := tx.QueryRowContext(ctx, q, id).
		Scan(&raw.id, &raw.name, &raw.email, &raw.phoneNumber, &raw.password)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return domain.User{}, context.Cause(ctx)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, tracerr.CustomError(portout.ErrDataNotFound, tracerr.StackTrace(err))
		}

		otel.SpanLogError(span, err, "failed to find user")
		return domain.User{}, tracerr.Wrap(err)
	}

	resRoleMap := <-chanRoleRes
	if resRoleMap.err != nil {
		return domain.User{}, resRoleMap.err
	}

	return domain.NewUser(
		domid.UserID(raw.id),
		raw.name,
		raw.phoneNumber,
		raw.email,
		raw.password,
		resRoleMap.res[domid.UserID(raw.id)],
	), nil
}

// FindByEmail not safe for transactions
func (repo *User) FindByEmail(ctx context.Context, tx portout.DBTX, email string) (domain.User, error) {
	ctx, span := monitorings.Tracer().Start(ctx, "db.pg.User.findByEmail")
	defer span.End()

	monitorings.Logger().InfoContext(ctx, "input", slog.String("email", email))

	var raw struct {
		id          string
		name        string
		phoneNumber string
		email       string
		password    string
	}

	err := tx.QueryRowContext(
		ctx,
		`
SELECT
    id,
    name,
    email,
    phone_number,
    password
FROM
    users
WHERE
    email = $1
LIMIT
    1;
        `,
		email,
	).Scan(&raw.id, &raw.name, &raw.email, &raw.phoneNumber, &raw.password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, tracerr.CustomError(portout.ErrDataNotFound, tracerr.StackTrace(err))
		}

		otel.SpanLogError(span, err, "failed to find user")
		return domain.User{}, tracerr.Wrap(err)
	}

	roles, err := repo.r.GetByUserID(ctx, tx, domid.UserID(raw.id))
	if err != nil {
		return domain.User{}, err
	}

	return domain.NewUser(
		domid.UserID(raw.id),
		raw.name,
		raw.phoneNumber,
		raw.email,
		raw.password,
		roles,
	), nil
}

func NewUser(r *Role) *User {
	return &User{
		r: r,
	}
}
