package ports

import (
	"context"
	"database/sql"
)

type UnitOfWork interface {
	Begin(ctx context.Context) error
	
	Commit() error
	
	Rollback() error
	
	ProductRepository() ProductRepository
	
	OutboxRepository() OutboxRepository
	
	Transaction() *sql.Tx
}

