package port

import (
	"context"
	"database/sql"
)

var (
	ErrDataNotFound = sql.ErrNoRows
	ErrTxDone       = sql.ErrTxDone
)

type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}
