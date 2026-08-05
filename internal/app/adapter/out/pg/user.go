package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/mon"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/queryutil"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/port"
	getallopt "FaisalBudiono/go-boilerplate/internal/app/port/options/user/getall"
)

func NewUser() *User {
	return &User{}
}

type userRes struct {
	id        string
	name      string
	email     string
	createdAt time.Time
	updatedAt time.Time
}

type userWithPassRes struct {
	id        string
	name      string
	email     string
	password  string
	createdAt time.Time
	updatedAt time.Time
}

type User struct{}

func (u *User) FindByEmail(
	ctx context.Context, tx port.DBTX, email string,
) (domain.UserWithPassword, error) {
	ctx, span := mon.Tracer().Start(ctx, u.sName("find-by-email"))
	defer span.End()

	mon.Logger().InfoContext(
		ctx, "input", slog.String("email", email),
	)

	query := `
		SELECT
			u.id,
			u.name,
			u.email,
			u.password,
			u.created_at,
			u.updated_at
		FROM
			users AS u
		WHERE
			u.deleted_at IS NULL
			AND u.email = $1
		LIMIT
			1;
	`
	args := []any{email}

	mon.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	var raw userWithPassRes
	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&raw.id,
		&raw.name,
		&raw.email,
		&raw.password,
		&raw.createdAt,
		&raw.updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.UserWithPassword{}, port.ErrDataNotFound
		}

		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to find user by email"),
		)
		return domain.UserWithPassword{}, err
	}

	return u.mapWithPassDomain(raw), nil
}

func (u *User) GetMap(
	ctx context.Context, tx port.DBTX, ids []string,
) (map[string]domain.User, error) {
	ctx, span := mon.Tracer().Start(ctx, u.sName("get-map"))
	defer span.End()

	mon.Logger().DebugContext(ctx, "ids", slog.Any("ids", ids))

	res := map[string]domain.User{}
	if len(ids) == 0 {
		mon.Logger().WarnContext(ctx, "user ids is empty")
		return res, nil
	}

	query := fmt.Sprintf(
		`
		SELECT
			u.id,
			u.name,
			u.email,
			u.created_at,
			u.updated_at
		FROM
			users AS u
		WHERE
			u.deleted_at IS NULL
			AND u.id IN (%s)
	`,
		queryutil.ArgsPlaceholder(len(ids), 0),
	)
	args := []any{}

	for _, rawID := range ids {
		realID, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			mon.Logger().ErrorContext(
				ctx, "failed to parse userID", slog.String("userID", rawID),
			)
			realID = -1
		}
		args = append(args, realID)
	}

	mon.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to query user"),
		)
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			otelutil.SpanLogError(
				span, err,
				otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to close rows"),
			)
		}
	}()

	for rows.Next() {
		var raw userRes
		err = rows.Scan(
			&raw.id,
			&raw.name,
			&raw.email,
			&raw.createdAt,
			&raw.updatedAt,
		)
		if err != nil {
			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to scan user"),
			)
			return nil, err
		}

		res[raw.id] = u.mapDomain(raw)
	}

	return res, nil
}

func (u *User) GetPaginated(
	ctx context.Context, tx port.DBTX,
	page int64, perPage int64,
	opts ...getallopt.QueryOption,
) ([]domain.User, int64, error) {
	ctx, span := mon.Tracer().Start(ctx, u.sName("get-paginated"))
	defer span.End()

	qo := getallopt.NewQueryOpt()
	for _, opt := range opts {
		opt(qo)
	}

	mon.Logger().DebugContext(
		ctx, "input",
		slog.Int64("page", page),
		slog.Int64("perPage", perPage),
		slog.Any("opts", qo),
	)

	args := []any{}
	conditions := []string{"u.deleted_at IS NULL"}

	if len(qo.Roles) > 0 {
		cond := fmt.Sprintf(
			`
			EXISTS (
				SELECT
					1
				FROM
					user_roles AS ur
				WHERE
					ur.user_id = u.id
					AND ur.name IN (%s)
			)
			`,
			queryutil.ArgsPlaceholder(len(qo.Roles), len(args)),
		)
		conditions = append(conditions, cond)

		for _, role := range qo.Roles {
			args = append(args, role.String())
		}
	}

	offset := (page - 1) * perPage

	query := fmt.Sprintf(
		`
		SELECT
			u.id,
			u.name,
			u.email,
			u.created_at,
			u.updated_at
		FROM
			users AS u
		WHERE
			%s
		ORDER BY
			u.name ASC
		LIMIT
			%d
		OFFSET
			%d
	`,
		strings.Join(conditions, " AND "),
		perPage,
		offset,
	)

	mon.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to query user"),
		)
		return nil, 0, err
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

	res := []domain.User{}
	for rows.Next() {
		var raw userRes
		err = rows.Scan(
			&raw.id,
			&raw.name,
			&raw.email,
			&raw.createdAt,
			&raw.updatedAt,
		)
		if err != nil {
			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to scan user"),
			)
			return nil, 0, err
		}
		res = append(res, u.mapDomain(raw))
	}

	countQuery := fmt.Sprintf(
		`
		SELECT
			COUNT(1)
		FROM
			users AS u
		WHERE
			%s
		`,
		strings.Join(conditions, " AND "),
	)

	mon.Logger().DebugContext(
		ctx, "count query",
		slog.String("query", queryutil.Clean(countQuery)),
		slog.Any("args", args),
	)

	var total int64
	err = tx.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to counting user"),
		)
		return nil, 0, err
	}

	return res, total, nil
}

func (u *User) Insert(
	ctx context.Context, tx port.DBTX, data domain.UserData,
) (string, error) {
	ctx, span := mon.Tracer().Start(ctx, u.sName("insert"))
	defer span.End()

	mon.Logger().DebugContext(
		ctx, "input", slog.Any("data", data),
	)

	query := `
		INSERT INTO
			users (name, email, password)
		VALUES
			($1, $2, $3)
		RETURNING
			id
	`
	args := []any{data.Name, data.Email, data.HashedPassword}

	mon.Logger().DebugContext(
		ctx, "query", slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	var id string
	err := tx.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to insert user"),
		)
		return "", err
	}

	return id, nil
}

func (u *User) sName(s string) string {
	return "adapter.pg.user." + s
}

func (u *User) mapDomain(raw userRes) domain.User {
	return domain.NewUser(
		raw.id,
		raw.name,
		raw.email,
		raw.createdAt,
		raw.updatedAt,
	)
}

func (u *User) mapWithPassDomain(raw userWithPassRes) domain.UserWithPassword {
	return domain.NewUserWithPassword(
		raw.id,
		raw.name,
		raw.email,
		raw.password,
		raw.createdAt,
		raw.updatedAt,
	)
}
