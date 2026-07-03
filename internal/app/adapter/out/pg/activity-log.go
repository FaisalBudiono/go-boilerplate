package pg

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/core/util/queryutil"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/domain/actlog"
	"komdigi-immigration/internal/app/port"
	"komdigi-immigration/internal/app/port/options/activitylog/getall"
)

func NewActivityLog() *ActivityLog {
	return &ActivityLog{}
}

type activityLogRes struct {
	id          string
	userID      *string
	loginMethod *string
	loginID     *string
	action      string
	createdAt   time.Time
}

type activityLogMetaRes struct {
	activityLogID string
	name          string
	value         string
}

type ActivityLog struct{}

func (a *ActivityLog) AddMetadata(
	ctx context.Context, tx port.DBTX,
	id string,
	metadata map[actlog.MetaKey]string,
) error {
	ctx, span := monitoring.Tracer().Start(ctx, a.name("add-metadata"))
	defer span.End()

	monitoring.Logger().DebugContext(
		ctx, "input",
		slog.String("id", id),
		slog.Any("metadata", metadata),
	)

	if len(metadata) == 0 {
		monitoring.Logger().DebugContext(ctx, "metadata is empty")
		return nil
	}

	realActivityLogID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to parse id"),
		)
		return err
	}

	const totalRow = 3

	args := []any{}
	values := []string{}

	for mk, val := range metadata {
		values = append(
			values,
			fmt.Sprintf("(%s)", queryutil.ArgsPlaceholder(totalRow, len(args))),
		)
		args = append(args, realActivityLogID, mk.String(), val)
	}

	query := fmt.Sprintf(
		`
		INSERT INTO
			activity_log_metas (activity_log_id, name, value)
		VALUES
			%s;
	`,
		strings.Join(values, ", "),
	)

	monitoring.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to insert metadata"),
		)
		return err
	}

	return nil
}

func (a *ActivityLog) GetKeyMapByLogIDs(
	ctx context.Context, tx port.DBTX, logIDs []string,
) (map[string]map[actlog.MetaKey]string, error) {
	ctx, span := monitoring.Tracer().Start(ctx, a.name("get-key-map-by-log-ids"))
	defer span.End()

	monitoring.Logger().DebugContext(
		ctx, "input", slog.Any("logIDs", logIDs),
	)

	if len(logIDs) == 0 {
		monitoring.Logger().DebugContext(ctx, "logIDs is empty")
		return nil, nil
	}

	query := fmt.Sprintf(
		`
		SELECT
			alm.activity_log_id,
			alm.name,
			alm.value
		FROM
			activity_log_metas AS alm
		WHERE
			alm.deleted_at IS NULL
			AND alm.activity_log_id IN (%s)
		`,
		queryutil.ArgsPlaceholder(len(logIDs), 0),
	)
	args := []any{}

	for _, rawID := range logIDs {
		realID, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			monitoring.Logger().WarnContext(
				ctx, "failed to parse logID", slog.String("logID", rawID),
			)
			realID = -1
		}
		args = append(args, realID)
	}

	monitoring.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to query activity log metas"),
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

	res := make(map[string]map[actlog.MetaKey]string)
	for rows.Next() {
		var raw activityLogMetaRes
		err = rows.Scan(
			&raw.activityLogID,
			&raw.name,
			&raw.value,
		)
		if err != nil {
			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to scan activity log meta"),
			)
			return nil, err
		}

		metas, ok := res[raw.activityLogID]
		if !ok {
			metas = make(map[actlog.MetaKey]string)
		}
		metas[actlog.MetaKey(raw.name)] = raw.value

		res[raw.activityLogID] = metas
	}

	return res, nil
}

func (a *ActivityLog) GetPaginated(
	ctx context.Context,
	tx port.DBTX,
	page int64,
	perPage int64,
	opts ...getall.QueryOption,
) ([]domain.ActivityLog, int64, error) {
	ctx, span := monitoring.Tracer().Start(ctx, a.name("get-paginated"))
	defer span.End()

	qo := getall.NewQueryOpt()
	for _, opt := range opts {
		opt(qo)
	}

	monitoring.Logger().DebugContext(
		ctx, "input",
		slog.Int64("page", page),
		slog.Int64("perPage", perPage),
		slog.Any("opts", qo),
	)

	args := []any{}
	conditions := []string{"al.deleted_at IS NULL"}

	if len(qo.UserIDFlags) > 0 {
		conditions = append(
			conditions, fmt.Sprintf(
				"al.user_id IN (%s)",
				queryutil.ArgsPlaceholder(len(qo.UserIDFlags), len(args)),
			),
		)
	}
	for _, userID := range qo.UserIDFlags {
		realID, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			monitoring.Logger().ErrorContext(
				ctx, "failed to parse userID", slog.String("userID", userID),
			)
			realID = -1
		}
		args = append(args, realID)
	}

	if len(qo.ActionFlags) > 0 {
		conditions = append(
			conditions, fmt.Sprintf(
				"al.action IN (%s)",
				queryutil.ArgsPlaceholder(len(qo.ActionFlags), len(args)),
			),
		)
	}
	for _, action := range qo.ActionFlags {
		args = append(args, action.String())
	}

	if !qo.CreatedFrom.IsZero() {
		conditions = append(
			conditions, fmt.Sprintf(
				"al.created_at >= %s",
				queryutil.ArgsPlaceholder(1, len(args)),
			),
		)
		args = append(args, qo.CreatedFrom)
	}

	if !qo.CreatedTo.IsZero() {
		conditions = append(
			conditions, fmt.Sprintf(
				"al.created_at <= %s",
				queryutil.ArgsPlaceholder(1, len(args)),
			),
		)
		args = append(args, qo.CreatedTo)
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(
		`
		SELECT
			al.id,
			al.user_id,
			al.login_method,
			al.login_id,
			al.action,
			al.created_at
		FROM
			activity_logs AS al
		WHERE
			%s
		ORDER BY
			al.created_at DESC
		LIMIT
			%d
		OFFSET
			%d
	`,
		strings.Join(conditions, " AND "),
		perPage,
		offset,
	)

	monitoring.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to query activity logs"),
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

	res := make([]domain.ActivityLog, 0)
	for rows.Next() {
		var raw activityLogRes
		err = rows.Scan(
			&raw.id,
			&raw.userID,
			&raw.loginMethod,
			&raw.loginID,
			&raw.action,
			&raw.createdAt,
		)
		if err != nil {
			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to scan activity log"),
			)
			return nil, 0, err
		}
		res = append(res, a.mapDomainRes(raw))
	}

	countQuery := fmt.Sprintf(
		`
		SELECT
			COUNT(1)
		FROM
			activity_logs AS al
		WHERE
			%s
		`,
		strings.Join(conditions, " AND "),
	)

	monitoring.Logger().DebugContext(
		ctx, "count query",
		slog.String("query", queryutil.Clean(countQuery)),
		slog.Any("args", args),
	)

	var total int64
	err = tx.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to counting activity logs"),
		)
		return nil, 0, err
	}

	return res, total, nil
}

func (a *ActivityLog) Insert(
	ctx context.Context, tx port.DBTX, data domain.ActivityLogData,
) (string, error) {
	ctx, span := monitoring.Tracer().Start(ctx, a.name("insert"))
	defer span.End()

	monitoring.Logger().DebugContext(
		ctx, "input", slog.Any("data", data),
	)

	query := `
		INSERT INTO
			activity_logs (action, user_id, login_method, login_id)
		VALUES
			($1, $2, $3, $4)
		RETURNING id;
	`
	args := []any{data.Action, data.ActorID, data.LoginMethod, data.LoginID}

	monitoring.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	var id string
	err := tx.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to query activity log"),
		)
		return "", err
	}

	return id, nil
}

func (a *ActivityLog) mapDomainRes(raw activityLogRes) domain.ActivityLog {
	var loginMethod *domain.LoginMethod
	if raw.loginMethod != nil {
		loginMethod = new(domain.LoginMethod(*raw.loginMethod))
	}

	return domain.NewActivityLog(
		raw.id,
		raw.userID,
		loginMethod,
		raw.loginID,
		actlog.Action(raw.action),
		raw.createdAt,
	)
}

func (a *ActivityLog) name(s string) string {
	return fmt.Sprintf("adapter.pg.activity-log.%s", s)
}
