package pg

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/core/util/queryutil"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/port"
)

func NewRole() *Role {
	return &Role{}
}

type Role struct{}

type roleRes struct {
	userID string
	name   string
}

func (r *Role) GetMapByUserID(
	ctx context.Context, tx port.DBTX, userIDs []string,
) (map[string][]domain.Role, error) {
	ctx, span := monitoring.Tracer().Start(ctx, "adapter.pg.role.get-map-by-user-id")
	defer span.End()

	monitoring.Logger().InfoContext(
		ctx, "input", slog.Any("userIDs", userIDs),
	)

	res := make(map[string][]domain.Role)
	if len(userIDs) == 0 {
		monitoring.Logger().WarnContext(ctx, "user ids is empty")
		return res, nil
	}

	query := fmt.Sprintf(
		`
		SELECT
			ur.user_id,
			ur.name
		FROM
			user_roles AS ur
		WHERE
			ur.user_id IN (%s);
	`,
		queryutil.ArgsPlaceholder(len(userIDs), 0),
	)
	args := []any{}

	for _, userID := range userIDs {
		realID, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			monitoring.Logger().ErrorContext(
				ctx, "failed to parse userID", slog.String("userID", userID),
			)
			realID = -1
		}
		args = append(args, realID)
	}

	monitoring.Logger().DebugContext(
		ctx, "query", slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to query user roles"),
		)
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to close rows"),
			)
		}
	}()

	for rows.Next() {
		var raw roleRes
		err = rows.Scan(
			&raw.userID,
			&raw.name,
		)
		if err != nil {
			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to scan user role"),
			)
			return nil, err
		}

		_, ok := res[raw.userID]
		if !ok {
			res[raw.userID] = []domain.Role{domain.Role(raw.name)}
			continue
		}

		res[raw.userID] = append(res[raw.userID], domain.Role(raw.name))
	}

	return res, nil
}
