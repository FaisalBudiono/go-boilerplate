package pg

import (
	"FaisalBudiono/go-boilerplate/internal/db"
	"FaisalBudiono/go-boilerplate/internal/domain"
	"context"

	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type userRepo struct {
	tracer trace.Tracer

	r *roleRepo
}

type user struct {
	id          string
	name        string
	phoneNumber string
	email       string
	password    string
}

func (repo *userRepo) FindByID(ctx context.Context, tx db.DBTX, id string) (domain.User, error) {
	ctx, span := repo.tracer.Start(ctx, "postgres: findByID users")
	defer span.End()

	span.SetAttributes(attribute.String("input.id", id))

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
    id = $1
LIMIT
    1;
        `,
		id,
	).Scan(&u.id, &u.name, &u.email, &u.phoneNumber, &u.password)
	if err != nil {
		return domain.User{}, tracerr.Wrap(err)
	}

	roles, err := repo.r.RefetchedRoles(ctx, tx, u.id)
	if err != nil {
		return domain.User{}, err
	}

	return domain.NewUser(
		u.id,
		u.name,
		u.phoneNumber,
		u.email,
		u.password,
		roles,
	), nil
}

func (repo *userRepo) FindByEmail(ctx context.Context, tx db.DBTX, email string) (domain.User, error) {
	ctx, span := repo.tracer.Start(ctx, "postgres: findByEmail users")
	defer span.End()

	span.SetAttributes(attribute.String("input.email", email))

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

	roles, err := repo.r.RefetchedRoles(ctx, tx, u.id)
	if err != nil {
		return domain.User{}, err
	}

	return domain.NewUser(
		u.id,
		u.name,
		u.phoneNumber,
		u.email,
		u.password,
		roles,
	), nil
}

func NewUser(tracer trace.Tracer, r *roleRepo) *userRepo {
	return &userRepo{
		tracer: tracer,

		r: r,
	}
}
