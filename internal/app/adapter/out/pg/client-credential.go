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

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/core/util/queryutil"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/port"
	getallopt "komdigi-immigration/internal/app/port/options/clientcred/getall"
)

func NewClientCredential() *ClientCredential {
	return &ClientCredential{}
}

type clientCredRes struct {
	id        string
	userID    string
	clientID  string
	createdAt time.Time
	updatedAt time.Time
}

type clientCredSecretRes struct {
	id           string
	userID       string
	clientID     string
	clientSecret string
	createdAt    time.Time
	updatedAt    time.Time
}

type ClientCredential struct{}

func (c *ClientCredential) DeleteByClientID(
	ctx context.Context, tx port.DBTX, clientID string,
) error {
	ctx, span := monitoring.Tracer().Start(ctx, c.spanName("delete-by-client-id"))
	defer span.End()

	monitoring.Logger().DebugContext(
		ctx, "input", slog.String("clientID", clientID),
	)

	query := `
		UPDATE client_credentials AS cc
		SET
			deleted_at = now()
		WHERE
			cc.client_id = $1
	`
	args := []any{clientID}

	monitoring.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to delete client credential by clientID"),
		)
		return err
	}
	return nil
}

func (c *ClientCredential) FindSecretByClientID(
	ctx context.Context, tx port.DBTX, clientID string,
) (domain.ClientCredentialSecret, error) {
	ctx, span := monitoring.Tracer().Start(ctx, c.spanName("get-map-by-client-id"))
	defer span.End()

	monitoring.Logger().DebugContext(
		ctx, "input", slog.String("clientID", clientID),
	)

	query := `
		SELECT
			cc.id,
			cc.user_id,
			cc.client_id,
			cc.client_secret,
			cc.created_at,
			cc.updated_at
		FROM
			client_credentials AS cc
		WHERE
			cc.deleted_at IS NULL
			AND cc.client_id = $1
	`
	args := []any{clientID}

	monitoring.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	emptyVal := domain.ClientCredentialSecret{}

	var raw clientCredSecretRes
	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&raw.id,
		&raw.userID,
		&raw.clientID,
		&raw.clientSecret,
		&raw.createdAt,
		&raw.updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyVal, errors.Join(port.ErrDataNotFound, err)
		}

		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to query client credential secret"),
		)
		return emptyVal, err
	}

	return domain.NewClientCredentialSecret(
		raw.id,
		raw.userID,
		raw.clientID,
		raw.clientSecret,
		raw.createdAt,
		raw.updatedAt,
	), nil
}

func (c *ClientCredential) GetMapByClientID(
	ctx context.Context, tx port.DBTX, clientIDs []string,
) (map[string]domain.ClientCredential, error) {
	ctx, span := monitoring.Tracer().Start(ctx, c.spanName("get-map-by-client-id"))
	defer span.End()

	monitoring.Logger().DebugContext(
		ctx, "input", slog.Any("clientIDs", clientIDs),
	)

	if len(clientIDs) == 0 {
		monitoring.Logger().DebugContext(ctx, "clientIDs is empty")
		return nil, nil
	}

	query := fmt.Sprintf(
		`
		SELECT
			cc.id,
			cc.user_id,
			cc.client_id,
			cc.created_at,
			cc.updated_at
		FROM
			client_credentials AS cc
		WHERE
			cc.deleted_at IS NULL
			AND cc.client_id IN (%s)
	`,
		queryutil.ArgsPlaceholder(len(clientIDs), 0),
	)
	args := []any{}

	for _, clientID := range clientIDs {
		args = append(args, clientID)
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
			otelutil.WithMessage("failed to query client credentials"),
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

	res := make(map[string]domain.ClientCredential)
	for rows.Next() {
		var raw clientCredRes
		err = rows.Scan(
			&raw.id,
			&raw.userID,
			&raw.clientID,
			&raw.createdAt,
			&raw.updatedAt,
		)
		if err != nil {
			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to scan client credential"),
			)
			return nil, err
		}

		res[raw.clientID] = c.mapToDomain(raw)
	}

	return res, nil
}

func (c *ClientCredential) GetPaginated(
	ctx context.Context, tx port.DBTX,
	page int64, perPage int64, opts ...getallopt.QueryOption,
) ([]domain.ClientCredential, int64, error) {
	ctx, span := monitoring.Tracer().Start(ctx, c.spanName("get-paginated"))
	defer span.End()

	qo := getallopt.NewQueryOpt()
	for _, opt := range opts {
		opt(qo)
	}

	monitoring.Logger().DebugContext(
		ctx, "input", slog.Int64("page", page),
		slog.Int64("perPage", perPage),
		slog.Any("opts", qo),
	)

	conditions := []string{"cc.deleted_at IS NULL"}
	args := []any{}

	if qo.SearchClientID != "" {
		conditions = append(
			conditions, fmt.Sprintf(
				"cc.client_id ILIKE '%%' || %s || '%%'",
				queryutil.ArgsPlaceholder(1, len(args)),
			),
		)

		args = append(args, qo.SearchClientID)
	}

	if len(qo.UserIDs) > 0 {
		conditions = append(
			conditions, fmt.Sprintf(
				"cc.user_id IN (%s)",
				queryutil.ArgsPlaceholder(len(qo.UserIDs), len(args)),
			),
		)
		for _, raw := range qo.UserIDs {
			realID, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				monitoring.Logger().ErrorContext(
					ctx, "failed to parse userID", slog.String("userID", raw),
				)
				realID = -1
			}
			args = append(args, realID)
		}
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(
		`
		SELECT
			cc.id,
			cc.user_id,
			cc.client_id,
			cc.created_at,
			cc.updated_at
		FROM
			client_credentials AS cc
		WHERE
			%s
		ORDER BY
			cc.created_at DESC
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
			otelutil.WithMessage("failed to query client credentials"),
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

	res := make([]domain.ClientCredential, 0)
	for rows.Next() {
		var raw clientCredRes
		err = rows.Scan(
			&raw.id,
			&raw.userID,
			&raw.clientID,
			&raw.createdAt,
			&raw.updatedAt,
		)
		if err != nil {
			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed to scan client credential"),
			)
			return nil, 0, err
		}
		res = append(res, c.mapToDomain(raw))
	}

	countQuery := fmt.Sprintf(
		`
		SELECT
			COUNT(1)
		FROM
			client_credentials AS cc
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
			otelutil.WithMessage("failed to counting client credentials"),
		)
		return nil, 0, err
	}

	return res, total, nil
}

func (c *ClientCredential) Insert(ctx context.Context, tx port.DBTX, data domain.ClientCredentialData) error {
	ctx, span := monitoring.Tracer().Start(ctx, c.spanName("insert"))
	defer span.End()

	query := `
		INSERT INTO
			client_credentials (user_id, client_id, client_secret)
		VALUES
			($1, $2, $3);
	`
	args := []any{data.UserID, data.ClientID, data.HashedSecret}

	monitoring.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to insert client credential"),
		)
		return err
	}
	return nil
}

func (c *ClientCredential) mapToDomain(raw clientCredRes) domain.ClientCredential {
	return domain.ClientCredential{
		ID:        raw.id,
		UserID:    raw.userID,
		ClientID:  raw.clientID,
		CreatedAt: raw.createdAt,
		UpdatedAt: raw.updatedAt,
	}
}

func (c *ClientCredential) spanName(s string) string {
	return fmt.Sprintf("adapter.pg.client-credential.%s", s)
}
