package repository

import (
	"context"
	"database/sql"
	"product_service/products/internal/usecase/ports"
)

var _ ports.UnitOfWork = (*postgresUnitOfWork)(nil)

type RepositoryFactory interface {
	CreateProductRepository() ports.ProductRepository
	CreateOutboxRepository() ports.OutboxRepository
}

type postgresUnitOfWork struct {
	db            *sql.DB
	tx            *sql.Tx
	productRepo   ports.ProductRepository
	outboxRepo    ports.OutboxRepository
	inTransaction bool
	productStm    *PreparedStatements
	outboxStm     *PreparedStatements
	metrics       ports.MetricsCollector
}

func NewUnitOfWork(db *sql.DB, productStm, outboxStm *PreparedStatements) ports.UnitOfWork {
	return &postgresUnitOfWork{
		db:         db,
		productStm: productStm,
		outboxStm:  outboxStm,
	}
}

func NewUnitOfWorkWithMetrics(db *sql.DB, productStm, outboxStm *PreparedStatements, metrics ports.MetricsCollector) ports.UnitOfWork {
	return &postgresUnitOfWork{
		db:         db,
		productStm: productStm,
		outboxStm:  outboxStm,
		metrics:    metrics,
	}
}

type uowFactory struct {
	db         *sql.DB
	productStm *PreparedStatements
	outboxStm  *PreparedStatements
	metrics    ports.MetricsCollector
}

var _ ports.UoWFactory = (*uowFactory)(nil)

func NewUoWFactory(db *sql.DB, productStm, outboxStm *PreparedStatements, metrics ports.MetricsCollector) ports.UoWFactory {
	return &uowFactory{
		db:         db,
		productStm: productStm,
		outboxStm:  outboxStm,
		metrics:    metrics,
	}
}

func (f *uowFactory) CreateUnitOfWork() ports.UnitOfWork {
	if f.metrics != nil {
		return NewUnitOfWorkWithMetrics(f.db, f.productStm, f.outboxStm, f.metrics)
	}
	return NewUnitOfWork(f.db, f.productStm, f.outboxStm)
}

func (u *postgresUnitOfWork) Begin(ctx context.Context) error {
	if u.inTransaction {
		return nil
	}
	
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	
	u.tx = tx
	u.inTransaction = true
	
	baseProductRepo := NewPostgresProductRepository(u.db, u.productStm)
	
	if u.metrics != nil {
		u.productRepo = NewMetricsProductRepositoryDecorator(baseProductRepo, u.metrics)
	} else {
		u.productRepo = baseProductRepo
	}
	
	u.outboxRepo = NewPostgresOutboxRepository(u.db, u.outboxStm)
	
	if txRepo, ok := u.productRepo.(ports.TransactionalRepository); ok {
		txRepo.SetTransaction(tx)
	}
	if txRepo, ok := u.outboxRepo.(ports.TransactionalRepository); ok {
		txRepo.SetTransaction(tx)
	}
	
	return nil
}

func (u *postgresUnitOfWork) Commit() error {
	if !u.inTransaction {
		return nil
	}
	
	if u.tx == nil {
		return nil
	}
	
	err := u.tx.Commit()
	u.inTransaction = false
	u.tx = nil
	
	if txRepo, ok := u.productRepo.(ports.TransactionalRepository); ok {
		txRepo.ClearTransaction()
	}
	if txRepo, ok := u.outboxRepo.(ports.TransactionalRepository); ok {
		txRepo.ClearTransaction()
	}
	
	return err
}

func (u *postgresUnitOfWork) Rollback() error {
	if !u.inTransaction {
		return nil
	}
	
	if u.tx == nil {
		return nil
	}
	
	err := u.tx.Rollback()
	u.inTransaction = false
	u.tx = nil
	
	if txRepo, ok := u.productRepo.(ports.TransactionalRepository); ok {
		txRepo.ClearTransaction()
	}
	if txRepo, ok := u.outboxRepo.(ports.TransactionalRepository); ok {
		txRepo.ClearTransaction()
	}
	
	return err
}

func (u *postgresUnitOfWork) ProductRepository() ports.ProductRepository {
	return u.productRepo
}

func (u *postgresUnitOfWork) OutboxRepository() ports.OutboxRepository {
	return u.outboxRepo
}

func (u *postgresUnitOfWork) Transaction() *sql.Tx {
	return u.tx
}

