package pg

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/queryutil"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout"
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
)

type Role struct{}

// ByUserIDs will return map[userID][]Role from slice of userIDs
func (repo *Role) ByUserIDs(
	ctx context.Context, tx portout.DBTX, ids []string,
) (map[string][]domain.Role, error) {
	ctx, span := monitoring.Tracer().Start(ctx, "db.pg.Role.ByUserIDs")
	defer span.End()

	if len(ids) == 0 {
		return make(map[string][]domain.Role), nil
	}

	monitoring.Logger().InfoContext(ctx, "input",
		slog.Any("ids", ids),
	)

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

	monitoring.Logger().DebugContext(ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to query")

		monitoring.Logger().ErrorContext(ctx, "failed to query",
			slog.Any("error", err),
		)

		return nil, err
	}
	defer rows.Close()

	rolesMap := make(map[string][]domain.Role)
	for rows.Next() {
		var userID, role string

		err = rows.Scan(&userID, &role)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to scan row")

			monitoring.Logger().ErrorContext(ctx, "failed to scan row",
				slog.Any("error", err),
			)

			return nil, err
		}

		rolesMap[userID] = append(rolesMap[userID], domain.Role(role))
	}

	return rolesMap, nil
}

func (repo *Role) GetByUserID(
	ctx context.Context, tx portout.DBTX, id string,
) ([]domain.Role, error) {
	ctx, span := monitoring.Tracer().Start(ctx, "db.pg.Role.GetByUserID")
	defer span.End()

	monitoring.Logger().InfoContext(ctx, "input", slog.Any("id", id))

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
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to query")

		monitoring.Logger().ErrorContext(ctx, "failed to query",
			slog.Any("error", err),
		)

		return nil, err
	}
	defer rows.Close()

	roles := make([]domain.Role, 0)
	for rows.Next() {
		var role domain.Role

		err = rows.Scan(&role)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to scan row")

			monitoring.Logger().ErrorContext(ctx, "failed to scan row",
				slog.Any("error", err),
			)

			return nil, err
		}

		roles = append(roles, role)
	}

	return roles, nil
}

func NewRole() *Role {
	return &Role{}
}
