package pg

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/db"
	"context"

	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type roleRepo struct {
	tracer trace.Tracer
}

func (repo *roleRepo) RefetchedRoles(ctx context.Context, tx db.DBTX, userID string) ([]domain.Role, error) {
	ctx, span := repo.tracer.Start(ctx, "postgres: refetched roles")
	defer span.End()

	span.SetAttributes(attribute.String("input.user.id", userID))

	rows, err := tx.QueryContext(ctx, `
SELECT
    name
FROM
    user_roles
WHERE
    user_id = $1
    AND deleted_at IS NULL
ORDER BY
    created_at ASC;
    `,
		userID,
	)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	defer rows.Close()

	roles := make([]domain.Role, 0)

	for rows.Next() {
		var role string

		err = rows.Scan(&role)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}

		roles = append(roles, domain.Role(role))
	}

	return roles, nil
}

func NewRole(tracer trace.Tracer) *roleRepo {
	return &roleRepo{
		tracer: tracer,
	}
}
