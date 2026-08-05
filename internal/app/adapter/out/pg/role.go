package pg

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/mon"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/queryutil"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/port"
)

func NewRole() *Role {
	return &Role{}
}

type roleRes struct {
	userID string
	name   string
}

type Role struct{}

func (r *Role) AddByUserID(
	ctx context.Context, tx port.DBTX, userID string, roles []domain.Role,
) error {
	ctx, span := mon.Tracer().Start(ctx, r.sName("add-by-user-id"))
	defer span.End()

	mon.Logger().DebugContext(
		ctx, "input", slog.String("userID", userID),
		slog.Any("roles", roles),
	)

	if len(roles) == 0 {
		mon.Logger().WarnContext(ctx, "roles is empty")
		return nil
	}

	realID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		mon.Logger().ErrorContext(
			ctx, "failed to parse userID", slog.String("userID", userID),
		)
		return err
	}

	args := []any{}
	values := make([]string, len(roles))
	for i, role := range roles {
		values[i] = fmt.Sprintf("(%s)", queryutil.ArgsPlaceholder(2, len(args)))
		args = append(args, realID, role.String())
	}

	query := fmt.Sprintf(
		`
		INSERT INTO
			user_roles (user_id, name)
		VALUES
			%s
	`,
		strings.Join(values, ","),
	)

	mon.Logger().DebugContext(
		ctx, "query", slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to insert user roles"),
		)
		return err
	}

	return nil
}

func (r *Role) CleanByUserID(
	ctx context.Context, tx port.DBTX, userID string,
) error {
	ctx, span := mon.Tracer().Start(ctx, r.sName("clean-by-user-id"))
	defer span.End()

	mon.Logger().DebugContext(
		ctx, "input", slog.String("userID", userID),
	)

	realID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		mon.Logger().ErrorContext(
			ctx, "failed to parse userID", slog.String("userID", userID),
		)
		return err
	}

	query := `
		DELETE FROM
			user_roles
		WHERE
			user_id = $1
	`
	args := []any{realID}

	mon.Logger().DebugContext(
		ctx, "query", slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to delete user roles"),
		)
		return err
	}

	return nil
}

func (r *Role) GetMapByUserID(
	ctx context.Context, tx port.DBTX, userIDs []string,
) (map[string][]domain.Role, error) {
	ctx, span := mon.Tracer().Start(ctx, r.sName("get-map-by-user-id"))
	defer span.End()

	mon.Logger().DebugContext(
		ctx, "input", slog.Any("userIDs", userIDs),
	)

	res := make(map[string][]domain.Role)
	if len(userIDs) == 0 {
		mon.Logger().WarnContext(ctx, "user ids is empty")
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
			mon.Logger().ErrorContext(
				ctx, "failed to parse userID", slog.String("userID", userID),
			)
			realID = -1
		}
		args = append(args, realID)
	}

	mon.Logger().DebugContext(
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

func (r *Role) sName(s string) string {
	return "adapter.pg.role." + s
}
