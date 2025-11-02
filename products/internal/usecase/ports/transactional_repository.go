package ports

import "database/sql"

type TransactionalRepository interface {
	SetTransaction(tx *sql.Tx)
	ClearTransaction()
}

