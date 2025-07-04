package pg

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout"
	"context"
	"errors"
	"log/slog"

	"github.com/ztrue/tracerr"
)

type User struct {
	r *Role
}

type user struct {
	id          string
	name        string
	phoneNumber string
	email       string
	password    string
}

type resultRoleMap struct {
	res map[domid.UserID][]domain.Role
	err error
}

func (repo *User) FindByID(ctx context.Context, tx portout.DBTX, id domid.UserID) (domain.User, error) {
	ctx, span := monitorings.Tracer().Start(ctx, "db.pg.user.findByID")
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
	monitorings.Logger().InfoContext(ctx, "making query", slog.String("query", q))

	u := user{}
	err := tx.QueryRowContext(ctx, q, id).
		Scan(&u.id, &u.name, &u.email, &u.phoneNumber, &u.password)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return domain.User{}, context.Cause(ctx)
		}

		return domain.User{}, tracerr.Wrap(err)
	}

	resRoleMap := <-chanRoleRes
	if resRoleMap.err != nil {
		return domain.User{}, resRoleMap.err
	}

	return domain.NewUser(
		domid.UserID(u.id),
		u.name,
		u.phoneNumber,
		u.email,
		u.password,
		resRoleMap.res[domid.UserID(u.id)],
	), nil
}

func (repo *User) FindByEmail(ctx context.Context, tx portout.DBTX, email string) (domain.User, error) {
	ctx, span := monitorings.Tracer().Start(ctx, "db.pg.user.findByEmail")
	defer span.End()

	monitorings.Logger().InfoContext(ctx, "input", slog.String("email", email))

	u := user{}

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
	).Scan(&u.id, &u.name, &u.email, &u.phoneNumber, &u.password)
	if err != nil {
		return domain.User{}, tracerr.Wrap(err)
	}

	roles, err := repo.r.RefetchedRoles(ctx, tx, domid.UserID(u.id))
	if err != nil {
		return domain.User{}, err
	}

	return domain.NewUser(
		domid.UserID(u.id),
		u.name,
		u.phoneNumber,
		u.email,
		u.password,
		roles,
	), nil
}

func NewUser(r *Role) *User {
	return &User{
		r: r,
	}
}
