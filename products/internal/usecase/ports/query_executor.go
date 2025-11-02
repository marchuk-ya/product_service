package ports

import (
	"context"
	"database/sql"
)

type QueryExecutor interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

var _ QueryExecutor = (*sql.DB)(nil)

var _ QueryExecutor = (*sql.Tx)(nil)

