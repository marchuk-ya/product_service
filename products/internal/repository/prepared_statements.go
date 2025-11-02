package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PreparedStatements struct {
	CreateProduct    *sql.Stmt
	GetProductByID   *sql.Stmt
	ListProducts     *sql.Stmt
	DeleteProduct    *sql.Stmt

	SaveOutboxEvent       *sql.Stmt
	GetPendingEvents      *sql.Stmt
	MarkAsPublished       *sql.Stmt
	MarkAsFailed          *sql.Stmt
	CheckIdempotencyKey   *sql.Stmt
}

func PrepareProductStatements(ctx context.Context, db *sql.DB) (*PreparedStatements, error) {
	createProduct, err := db.PrepareContext(ctx, queryCreateProduct)
	if err != nil {
		return nil, err
	}

	getProductByID, err := db.PrepareContext(ctx, queryGetProductByID)
	if err != nil {
		return nil, err
	}

	listProducts, err := db.PrepareContext(ctx, queryListProducts)
	if err != nil {
		return nil, err
	}

	deleteProduct, err := db.PrepareContext(ctx, queryDeleteProduct)
	if err != nil {
		return nil, err
	}

	return &PreparedStatements{
		CreateProduct:  createProduct,
		GetProductByID: getProductByID,
		ListProducts:   listProducts,
		DeleteProduct:  deleteProduct,
	}, nil
}

func PrepareOutboxStatements(ctx context.Context, db *sql.DB) (*PreparedStatements, error) {
	saveOutboxEvent, err := db.PrepareContext(ctx, querySaveOutboxEvent)
	if err != nil {
		return nil, err
	}

	getPendingEvents, err := db.PrepareContext(ctx, queryGetPendingEvents)
	if err != nil {
		return nil, err
	}

	markAsPublished, err := db.PrepareContext(ctx, queryMarkAsPublished)
	if err != nil {
		return nil, err
	}

	markAsFailed, err := db.PrepareContext(ctx, queryMarkAsFailed)
	if err != nil {
		return nil, err
	}

	checkIdempotencyKey, err := db.PrepareContext(ctx, queryCheckIdempotencyKey)
	if err != nil {
		return nil, err
	}

	return &PreparedStatements{
		SaveOutboxEvent:     saveOutboxEvent,
		GetPendingEvents:    getPendingEvents,
		MarkAsPublished:     markAsPublished,
		MarkAsFailed:        markAsFailed,
		CheckIdempotencyKey: checkIdempotencyKey,
	}, nil
}

func (ps *PreparedStatements) Close() error {
	var errs []error
	if ps.CreateProduct != nil {
		if e := ps.CreateProduct.Close(); e != nil {
			errs = append(errs, fmt.Errorf("CreateProduct: %w", e))
		}
	}
	if ps.GetProductByID != nil {
		if e := ps.GetProductByID.Close(); e != nil {
			errs = append(errs, fmt.Errorf("GetProductByID: %w", e))
		}
	}
	if ps.ListProducts != nil {
		if e := ps.ListProducts.Close(); e != nil {
			errs = append(errs, fmt.Errorf("ListProducts: %w", e))
		}
	}
	if ps.DeleteProduct != nil {
		if e := ps.DeleteProduct.Close(); e != nil {
			errs = append(errs, fmt.Errorf("DeleteProduct: %w", e))
		}
	}
	if ps.SaveOutboxEvent != nil {
		if e := ps.SaveOutboxEvent.Close(); e != nil {
			errs = append(errs, fmt.Errorf("SaveOutboxEvent: %w", e))
		}
	}
	if ps.GetPendingEvents != nil {
		if e := ps.GetPendingEvents.Close(); e != nil {
			errs = append(errs, fmt.Errorf("GetPendingEvents: %w", e))
		}
	}
	if ps.MarkAsPublished != nil {
		if e := ps.MarkAsPublished.Close(); e != nil {
			errs = append(errs, fmt.Errorf("MarkAsPublished: %w", e))
		}
	}
	if ps.MarkAsFailed != nil {
		if e := ps.MarkAsFailed.Close(); e != nil {
			errs = append(errs, fmt.Errorf("MarkAsFailed: %w", e))
		}
	}
	if ps.CheckIdempotencyKey != nil {
		if e := ps.CheckIdempotencyKey.Close(); e != nil {
			errs = append(errs, fmt.Errorf("CheckIdempotencyKey: %w", e))
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("failed to close %d prepared statements: %w", len(errs), errors.Join(errs...))
	}
	return nil
}


