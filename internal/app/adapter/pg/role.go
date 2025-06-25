package pg

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel/spanattr"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/queryutil"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout"
	"context"
	"fmt"

	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/attribute"
)

type roleRepo struct{}

func (repo *roleRepo) ByUserIDs(ctx context.Context, tx portout.DBTX, ids []domid.UserID) (map[domid.UserID][]domain.Role, error) {
	ctx, span := monitorings.Tracer().Start(ctx, "postgres: refetched roles")
	defer span.End()

	if len(ids) == 0 {
		return nil, tracerr.New("User IDs is empty")
	}

	span.SetAttributes(attribute.String("input.user.id", fmt.Sprintf("%#v", ids)))

	query := fmt.Sprintf(
		`
SELECT
    user_id,
    name
FROM
    user_roles
WHERE
    deleted_at IS NULL
    AND user_id IN (%s)
ORDER BY
    created_at ASC;
`,
		queryutil.ArgsPlaceholder(len(ids), 0),
	)

	args := make([]any, len(ids))
	for i := range ids {
		args[i] = ids[i]
	}

	span.SetAttributes(
		spanattr.Query(query),
		attribute.String("query.args", fmt.Sprintf("%#v", args)),
	)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	defer rows.Close()

	rolesMap := make(map[domid.UserID][]domain.Role)

	for rows.Next() {
		var userID, role string

		err = rows.Scan(&userID, &role)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}

		rolesMap[domid.UserID(userID)] = append(rolesMap[domid.UserID(userID)], domain.Role(role))
	}

	return rolesMap, nil
}

func (repo *roleRepo) RefetchedRoles(ctx context.Context, tx portout.DBTX, id domid.UserID) ([]domain.Role, error) {
	ctx, span := monitorings.Tracer().Start(ctx, "postgres: refetched roles")
	defer span.End()

	span.SetAttributes(attribute.String("input.user.id", string(id)))

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
		id,
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

func NewRole() *roleRepo {
	return &roleRepo{}
}
