package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"product_service/products/internal/domain"
	"product_service/products/internal/usecase/ports"
)

var _ ports.ProductRepository = (*postgresProductRepository)(nil)

type postgresProductRepository struct {
	db            *sql.DB
	tx            *sql.Tx
	stm           *PreparedStatements
	queryExecutor ports.QueryExecutor
}

func NewPostgresProductRepository(db *sql.DB, stm *PreparedStatements) ports.ProductRepository {
	return &postgresProductRepository{
		db:  db,
		stm: stm,
	}
}

func (r *postgresProductRepository) getQueryExecutor() ports.QueryExecutor {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *postgresProductRepository) SetTransaction(tx *sql.Tx) {
	r.tx = tx
	r.queryExecutor = tx
}

func (r *postgresProductRepository) ClearTransaction() {
	r.tx = nil
	r.queryExecutor = r.db
}

var _ ports.TransactionalRepository = (*postgresProductRepository)(nil)

func (r *postgresProductRepository) Create(ctx context.Context, product *domain.Product) error {
	var row *sql.Row

	if r.tx != nil {
		txStmt := r.tx.StmtContext(ctx, r.stm.CreateProduct)
		defer txStmt.Close()
		row = txStmt.QueryRowContext(ctx, product.Name.Value(), product.Price.Value())
	} else {
		row = r.stm.CreateProduct.QueryRowContext(ctx, product.Name.Value(), product.Price.Value())
	}

	err := row.Scan(&product.ID, &product.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

func (r *postgresProductRepository) GetByID(ctx context.Context, id int) (*domain.Product, error) {
	product := &domain.Product{}
	var name string
	var price float64

	var row *sql.Row
	if r.tx != nil {
		txStmt := r.tx.StmtContext(ctx, r.stm.GetProductByID)
		defer txStmt.Close()
		row = txStmt.QueryRowContext(ctx, id)
	} else {
		row = r.stm.GetProductByID.QueryRowContext(ctx, id)
	}

	err := row.Scan(&product.ID, &name, &price, &product.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("product not found: %w", domain.ErrProductNotFound)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	productName, err := domain.NewProductName(name)
	if err != nil {
		return nil, fmt.Errorf("invalid product data: %w", err)
	}
	product.Name = productName

	productPrice, err := domain.NewPrice(price)
	if err != nil {
		return nil, fmt.Errorf("invalid product data: %w", err)
	}
	product.Price = productPrice

	return product, nil
}

func (r *postgresProductRepository) executeQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, func() error, error) {
	var rows *sql.Rows
	var closeFn func() error = func() error { return nil }
	var err error

	if r.tx != nil {
		txStmt := r.tx.StmtContext(ctx, r.stm.ListProducts)
		rows, err = txStmt.QueryContext(ctx, args...)
		if err != nil {
			txStmt.Close()
			return nil, nil, fmt.Errorf("failed to execute query: %w", err)
		}
		closeFn = txStmt.Close
	} else {
		rows, err = r.stm.ListProducts.QueryContext(ctx, args...)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return rows, closeFn, nil
}

func (r *postgresProductRepository) List(ctx context.Context, page, limit int) ([]domain.Product, int, error) {
	offset := (page - 1) * limit

	rows, closeFn, err := r.executeQuery(ctx, queryListProducts, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if err := closeFn(); err != nil {
		}
		if err := rows.Close(); err != nil {
		}
	}()

	products := make([]domain.Product, 0, limit)
	var total int

	for rows.Next() {
		var product domain.Product
		var name string
		var price float64

		if err := rows.Scan(&product.ID, &name, &price, &product.CreatedAt, &total); err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		productName, err := domain.NewProductName(name)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid product data: %w", err)
		}
		product.Name = productName

		productPrice, err := domain.NewPrice(price)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid product data: %w", err)
		}
		product.Price = productPrice

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate products: %w", err)
	}

	return products, total, nil
}

func (r *postgresProductRepository) executeExec(ctx context.Context, args ...interface{}) (sql.Result, func() error, error) {
	var result sql.Result
	var closeFn func() error = func() error { return nil }
	var err error

	if r.tx != nil {
		txStmt := r.tx.StmtContext(ctx, r.stm.DeleteProduct)
		result, err = txStmt.ExecContext(ctx, args...)
		if err != nil {
			txStmt.Close()
			return nil, nil, fmt.Errorf("failed to execute query: %w", err)
		}
		closeFn = txStmt.Close
	} else {
		result, err = r.stm.DeleteProduct.ExecContext(ctx, args...)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return result, closeFn, nil
}

func (r *postgresProductRepository) Delete(ctx context.Context, id int) error {
	result, closeFn, err := r.executeExec(ctx, id)
	if err != nil {
		return err
	}
	defer closeFn()

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found: %w", domain.ErrProductNotFound)
	}

	return nil
}
