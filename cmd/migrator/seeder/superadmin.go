package seeder

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/hash"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"context"
	"database/sql"
	"errors"

	"github.com/spf13/viper"
	"github.com/ztrue/tracerr"
)

type superAdmin struct {
	ctx context.Context
	db  domain.DBTX
}

func (r *superAdmin) Name() string {
	return "first admin"
}

func (r *superAdmin) Seed() error {
	hasher := hash.NewArgon()
	password, err := hasher.Generate(viper.GetString("SEEDER_FIRST_ADMIN_PASSWORD"))
	if err != nil {
		return err
	}

	var foundId int64
	err = r.db.QueryRowContext(
		r.ctx,
		`
UPDATE users
SET
    name = $1,
    email = $2,
    phone_number = $3,
    password = $4,
    deleted_at = NULL
WHERE
    id = 1
RETURNING id
`,
		viper.GetString("SEEDER_FIRST_ADMIN_NAME"),
		viper.GetString("SEEDER_FIRST_ADMIN_EMAIL"),
		viper.GetString("SEEDER_FIRST_ADMIN_PHONE_NUMBER"),
		password,
	).Scan(&foundId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return tracerr.Wrap(err)
		}

		err = r.insertUser()
		if err != nil {
			return err
		}
	}

	var foundRoleID int64
	err = r.db.QueryRowContext(
		r.ctx,
		`
SELECT id
FROM user_roles 
WHERE
    user_id = 1
    AND deleted_at IS NULL
    AND name = $1
LIMIT 1
`,
		domain.RoleAdmin,
	).Scan(&foundRoleID)
	if err == nil {
		return nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return tracerr.Wrap(err)
	}

	return r.insertAdminRole()
}

func (r *superAdmin) insertUser() error {
	hasher := hash.NewArgon()
	password, err := hasher.Generate(viper.GetString("SEEDER_FIRST_ADMIN_PASSWORD"))
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(
		r.ctx,
		`
INSERT INTO
    users (id, name, email, phone_number, password, deleted_at)
VALUES
    (1, $1, $2, $3, $4, NULL)
`,
		viper.GetString("SEEDER_FIRST_ADMIN_NAME"),
		viper.GetString("SEEDER_FIRST_ADMIN_EMAIL"),
		viper.GetString("SEEDER_FIRST_ADMIN_PHONE_NUMBER"),
		password,
	)

	return tracerr.Wrap(err)
}

func (r *superAdmin) insertAdminRole() error {
	_, err := r.db.ExecContext(
		r.ctx,
		`
INSERT INTO
    user_roles (user_id, name)
VALUES
    (1, $1)
`,
		domain.RoleAdmin,
	)

	return tracerr.Wrap(err)
}

func NewSuperAdmin(ctx context.Context, db domain.DBTX) *superAdmin {
	return &superAdmin{ctx, db}
}
