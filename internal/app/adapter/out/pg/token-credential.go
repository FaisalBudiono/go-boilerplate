package pg

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/mon"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/queryutil"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/port"
)

func NewTokenCredential() *TokenCredential {
	return &TokenCredential{}
}

type tokenCredRes struct {
	id           string
	userID       string
	clientID     string
	hashedSecret string
	loginMethod  string
	loginID      string
}

type TokenCredential struct{}

func (t *TokenCredential) Insert(
	ctx context.Context, tx port.DBTX, data domain.TokenCredentialData,
) error {
	ctx, span := mon.Tracer().Start(ctx, "adapter.pg.token-credential.insert")
	defer span.End()

	query := `
		INSERT INTO
			token_credentials (user_id, client_id, client_secret, login_method, login_id)
		VALUES
			($1, $2, $3, $4, $5);
	`
	args := []any{
		data.UserID,
		data.ClientID,
		data.HashedSecret,
		string(data.LoginMethod),
		data.LoginID,
	}

	mon.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to insert token credential"),
		)
		return err
	}
	return nil
}

func (t *TokenCredential) DeleteByClientID(
	ctx context.Context, tx port.DBTX, clientID string,
) error {
	ctx, span := mon.Tracer().Start(ctx, "adapter.pg.token-credential.delete-by-client-id")
	defer span.End()

	query := `
		DELETE FROM
			token_credentials AS tc
		WHERE
			tc.client_id = $1
	`
	args := []any{clientID}

	mon.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to delete token credential"),
		)
		return err
	}

	return nil
}

func (t *TokenCredential) FindByClientID(
	ctx context.Context, tx port.DBTX, clientID string,
) (domain.TokenCredential, error) {
	ctx, span := mon.Tracer().Start(ctx, "adapter.pg.token-credential.find-by-client-id")
	defer span.End()

	query := `
		SELECT
			tc.id,
			tc.user_id,
			tc.client_id,
			tc.client_secret,
			tc.login_id,
			tc.login_method
		FROM
			token_credentials AS tc
		WHERE
			tc.client_id = $1
		LIMIT
			1
	`
	args := []any{clientID}

	mon.Logger().DebugContext(
		ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Any("args", args),
	)

	var raw tokenCredRes
	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&raw.id,
		&raw.userID,
		&raw.clientID,
		&raw.hashedSecret,
		&raw.loginMethod,
		&raw.loginID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			mon.Logger().DebugContext(
				ctx, "clientID not found", slog.Any("err", err),
			)
			return domain.TokenCredential{}, errors.Join(port.ErrDataNotFound, err)
		}

		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to find token credential by client ID"),
		)
		return domain.TokenCredential{}, err
	}

	return t.mapToDomain(raw), nil
}

func (t *TokenCredential) mapToDomain(raw tokenCredRes) domain.TokenCredential {
	return domain.TokenCredential{
		ID:           raw.id,
		UserID:       raw.userID,
		ClientID:     raw.clientID,
		HashedSecret: raw.hashedSecret,
		LoginMethod:  domain.LoginMethod(raw.loginMethod),
		LoginID:      raw.loginID,
	}
}
