package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"product_service/products/internal/usecase/ports"
	"strconv"
	"strings"
)

var _ ports.OutboxRepository = (*postgresOutboxRepository)(nil)

var _ ports.BatchOutboxRepository = (*postgresOutboxRepository)(nil)

type postgresOutboxRepository struct {
	db            *sql.DB
	tx            *sql.Tx
	stm           *PreparedStatements
	queryExecutor ports.QueryExecutor
	maxBatchSize  int
}

func NewPostgresOutboxRepository(db *sql.DB, stm *PreparedStatements) ports.OutboxRepository {
	return NewPostgresOutboxRepositoryWithConfig(db, stm, DefaultMaxBatchSize)
}

func NewPostgresOutboxRepositoryWithConfig(db *sql.DB, stm *PreparedStatements, maxBatchSize int) ports.OutboxRepository {
	return &postgresOutboxRepository{
		db:            db,
		stm:           stm,
		queryExecutor: db,
		maxBatchSize:  maxBatchSize,
	}
}

func (r *postgresOutboxRepository) getQueryExecutor() ports.QueryExecutor {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *postgresOutboxRepository) SetTransaction(tx *sql.Tx) {
	r.tx = tx
	r.queryExecutor = tx
}

func (r *postgresOutboxRepository) ClearTransaction() {
	r.tx = nil
	r.queryExecutor = r.db
}

var _ ports.TransactionalRepository = (*postgresOutboxRepository)(nil)

func (r *postgresOutboxRepository) executeQueryRow(ctx context.Context, query string, args ...interface{}) (*sql.Row, func() error, error) {
	executor := r.getQueryExecutor()
	var row *sql.Row
	var closeFn func() error = func() error { return nil }
	var stmt *sql.Stmt

	if r.tx != nil {
		var prepErr error
		stmt, prepErr = executor.PrepareContext(ctx, query)
		if prepErr != nil {
			return nil, nil, fmt.Errorf("failed to prepare statement: %w", prepErr)
		}
		row = stmt.QueryRowContext(ctx, args...)
		closeFn = func() error {
			if stmt != nil {
				return stmt.Close()
			}
			return nil
		}
	} else {
		row = executor.QueryRowContext(ctx, query, args...)
		closeFn = func() error { return nil }
	}

	return row, closeFn, nil
}

func (r *postgresOutboxRepository) executeQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, func() error, error) {
	executor := r.getQueryExecutor()
	var rows *sql.Rows
	var closeFn func() error = func() error { return nil }
	var err error
	var stmt *sql.Stmt

	if r.tx != nil {
		var prepErr error
		stmt, prepErr = executor.PrepareContext(ctx, query)
		if prepErr != nil {
			return nil, nil, fmt.Errorf("failed to prepare statement: %w", prepErr)
		}
		rows, err = stmt.QueryContext(ctx, args...)
		if err != nil {
			if stmt != nil {
				stmt.Close()
			}
			return nil, nil, fmt.Errorf("failed to execute query: %w", err)
		}
		closeFn = func() error {
			var errs []error
			if rows != nil {
				if err := rows.Close(); err != nil {
					errs = append(errs, fmt.Errorf("failed to close rows: %w", err))
				}
			}
			if stmt != nil {
				if err := stmt.Close(); err != nil {
					errs = append(errs, fmt.Errorf("failed to close statement: %w", err))
				}
			}
			if len(errs) > 0 {
				return fmt.Errorf("errors closing resources: %v", errors.Join(errs...))
			}
			return nil
		}
	} else {
		rows, err = executor.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to execute query: %w", err)
		}
		closeFn = func() error {
			if rows != nil {
				return rows.Close()
			}
			return nil
		}
	}

	return rows, closeFn, nil
}

func (r *postgresOutboxRepository) executeExec(ctx context.Context, query string, args ...interface{}) (sql.Result, func() error, error) {
	executor := r.getQueryExecutor()
	var result sql.Result
	var closeFn func() error = func() error { return nil }
	var err error
	var stmt *sql.Stmt

	if r.tx != nil {
		var prepErr error
		stmt, prepErr = executor.PrepareContext(ctx, query)
		if prepErr != nil {
			return nil, nil, fmt.Errorf("failed to prepare statement: %w", prepErr)
		}
		result, err = stmt.ExecContext(ctx, args...)
		if err != nil {
			if stmt != nil {
				stmt.Close()
			}
			return nil, nil, fmt.Errorf("failed to execute query: %w", err)
		}
		closeFn = func() error {
			if stmt != nil {
				return stmt.Close()
			}
			return nil
		}
	} else {
		result, err = executor.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to execute query: %w", err)
		}
		closeFn = func() error { return nil }
	}

	return result, closeFn, nil
}

func toPortsOutboxEvent(event *OutboxEvent) *ports.OutboxEvent {
	var eventData []byte
	if event.EventData != nil {
		eventData = []byte(event.EventData)
	}
	return &ports.OutboxEvent{
		ID:             event.ID,
		EventType:      event.EventType,
		EventData:      eventData,
		IdempotencyKey: event.IdempotencyKey,
		CreatedAt:      event.CreatedAt,
		PublishedAt:    event.PublishedAt,
		RetryCount:     event.RetryCount,
		Status:         ports.OutboxStatus(event.Status),
	}
}

func fromPortsOutboxEvent(event *ports.OutboxEvent) *OutboxEvent {
	return &OutboxEvent{
		ID:             event.ID,
		EventType:      event.EventType,
		EventData:      json.RawMessage(event.EventData),
		IdempotencyKey: event.IdempotencyKey,
		CreatedAt:      event.CreatedAt,
		PublishedAt:    event.PublishedAt,
		RetryCount:     event.RetryCount,
		Status:         OutboxStatus(event.Status),
	}
}

func (r *postgresOutboxRepository) SaveEvent(ctx context.Context, event *ports.OutboxEvent) error {
	repoEvent := fromPortsOutboxEvent(event)

	var eventDataJSON json.RawMessage
	if repoEvent.EventData != nil {
		eventDataJSON = repoEvent.EventData
	} else {
		eventDataJSON = []byte("{}")
	}

	var row *sql.Row

	if r.tx != nil {
		txStmt := r.tx.StmtContext(ctx, r.stm.SaveOutboxEvent)
		defer func() {
			if err := txStmt.Close(); err != nil {
			}
		}()
		row = txStmt.QueryRowContext(ctx,
			repoEvent.EventType,
			eventDataJSON,
			repoEvent.IdempotencyKey,
			repoEvent.Status,
		)
	} else {
		row = r.stm.SaveOutboxEvent.QueryRowContext(ctx,
			repoEvent.EventType,
			eventDataJSON,
			repoEvent.IdempotencyKey,
			repoEvent.Status,
		)
	}

	err := row.Scan(&repoEvent.ID, &repoEvent.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) && repoEvent.IdempotencyKey != "" {
			return nil
		}
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	event.ID = repoEvent.ID
	event.CreatedAt = repoEvent.CreatedAt

	return nil
}

func (r *postgresOutboxRepository) GetPendingEvents(ctx context.Context, limit int) ([]ports.OutboxEvent, error) {
	var rows *sql.Rows
	var closeFn func() error = func() error { return nil }
	var err error

	if r.tx != nil {
		txStmt := r.tx.StmtContext(ctx, r.stm.GetPendingEvents)
		rows, err = txStmt.QueryContext(ctx, OutboxStatusPending, limit)
		if err != nil {
			txStmt.Close()
			return nil, fmt.Errorf("failed to get pending events: %w", err)
		}
		closeFn = func() error {
			var errs []error
			if rows != nil {
				if err := rows.Close(); err != nil {
					errs = append(errs, fmt.Errorf("failed to close rows: %w", err))
				}
			}
			if err := txStmt.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close statement: %w", err))
			}
			if len(errs) > 0 {
				return fmt.Errorf("errors closing resources: %v", errors.Join(errs...))
			}
			return nil
		}
	} else {
		rows, err = r.stm.GetPendingEvents.QueryContext(ctx, OutboxStatusPending, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get pending events: %w", err)
		}
		closeFn = func() error {
			if rows != nil {
				return rows.Close()
			}
			return nil
		}
	}

	return r.scanEventsWithCapacity(rows, closeFn, limit)
}

func (r *postgresOutboxRepository) scanEvents(rows *sql.Rows, closeFn func() error) ([]ports.OutboxEvent, error) {
	return r.scanEventsWithCapacity(rows, closeFn, initialScanCapacity)
}

func (r *postgresOutboxRepository) scanEventsWithCapacity(rows *sql.Rows, closeFn func() error, capacity int) ([]ports.OutboxEvent, error) {
	defer func() {
		if err := closeFn(); err != nil {
		}
		if err := rows.Close(); err != nil {
		}
	}()

	events := make([]ports.OutboxEvent, 0, capacity)
	for rows.Next() {
		var event OutboxEvent
		var publishedAt sql.NullTime
		var idempotencyKey sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.EventData,
			&idempotencyKey,
			&event.CreatedAt,
			&publishedAt,
			&event.RetryCount,
			&event.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan outbox event: %w", err)
		}

		if publishedAt.Valid {
			event.PublishedAt = &publishedAt.Time
		}
		if idempotencyKey.Valid {
			event.IdempotencyKey = idempotencyKey.String
		}

		events = append(events, *toPortsOutboxEvent(&event))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pending events: %w", err)
	}

	return events, nil
}

func (r *postgresOutboxRepository) MarkAsPublished(ctx context.Context, eventID int64) error {
	var result sql.Result
	var err error

	if r.tx != nil {
		txStmt := r.tx.StmtContext(ctx, r.stm.MarkAsPublished)
		defer txStmt.Close()
		result, err = txStmt.ExecContext(ctx, string(OutboxStatusPublished), eventID)
	} else {
		result, err = r.stm.MarkAsPublished.ExecContext(ctx, string(OutboxStatusPublished), eventID)
	}

	if err != nil {
		return fmt.Errorf("failed to mark event as published: %w", err)
	}

	return r.checkRowsAffected(result, eventID, "mark as published")
}

func (r *postgresOutboxRepository) checkRowsAffected(result sql.Result, eventID int64, operation string) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("event with id %d not found", eventID)
	}

	return nil
}

func (r *postgresOutboxRepository) MarkAsFailed(ctx context.Context, eventID int64, retryCount int) error {
	var result sql.Result
	var err error

	if r.tx != nil {
		txStmt := r.tx.StmtContext(ctx, r.stm.MarkAsFailed)
		defer txStmt.Close()
		result, err = txStmt.ExecContext(ctx, string(OutboxStatusFailed), retryCount, eventID)
	} else {
		result, err = r.stm.MarkAsFailed.ExecContext(ctx, string(OutboxStatusFailed), retryCount, eventID)
	}

	if err != nil {
		return fmt.Errorf("failed to mark event as failed: %w", err)
	}

	return r.checkRowsAffected(result, eventID, "mark as failed")
}

func (r *postgresOutboxRepository) CheckIdempotencyKey(ctx context.Context, idempotencyKey string) (bool, error) {
	if idempotencyKey == "" {
		return false, nil
	}

	var row *sql.Row

	if r.tx != nil {
		txStmt := r.tx.StmtContext(ctx, r.stm.CheckIdempotencyKey)
		defer txStmt.Close()
		row = txStmt.QueryRowContext(ctx, idempotencyKey)
	} else {
		row = r.stm.CheckIdempotencyKey.QueryRowContext(ctx, idempotencyKey)
	}

	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check idempotency key: %w", err)
	}

	return exists, nil
}

func (r *postgresOutboxRepository) MoveToDLQ(ctx context.Context, eventID int64, reason string) error {
	result, closeFn, err := r.executeExec(ctx, queryMoveToDLQ, string(OutboxStatusDLQ), eventID, reason)
	if err != nil {
		return err
	}
	defer closeFn()

	return r.checkRowsAffected(result, eventID, "move to DLQ")
}

const DefaultMaxBatchSize = 100

const paramsPerEvent = 4

const initialScanCapacity = 32

func (r *postgresOutboxRepository) SaveEventsBatch(ctx context.Context, events []*ports.OutboxEvent) error {
	if len(events) == 0 {
		return nil
	}

	if len(events) > r.maxBatchSize {
		return r.saveEventsBatchInChunks(ctx, events)
	}

	return r.saveEventsBatchSingle(ctx, events)
}

func (r *postgresOutboxRepository) saveEventsBatchSingle(ctx context.Context, events []*ports.OutboxEvent) error {
	args := make([]interface{}, 0, len(events)*paramsPerEvent)

	estimatedSize := len(querySaveEventsBatch) + len(" RETURNING id, created_at") + (len(events) * 20)
	var queryBuilder strings.Builder
	queryBuilder.Grow(estimatedSize)
	queryBuilder.WriteString(querySaveEventsBatch)

	argIndex := 1
	for i, event := range events {
		var eventDataJSON json.RawMessage
		if event.EventData != nil {
			eventDataJSON = event.EventData
		} else {
			eventDataJSON = []byte("{}")
		}

		if i > 0 {
			queryBuilder.WriteString(", ")
		}

		queryBuilder.WriteByte('(')
		queryBuilder.WriteByte('$')
		writeInt(&queryBuilder, argIndex)
		queryBuilder.WriteString(", $")
		writeInt(&queryBuilder, argIndex+1)
		queryBuilder.WriteString(", $")
		writeInt(&queryBuilder, argIndex+2)
		queryBuilder.WriteString(", $")
		writeInt(&queryBuilder, argIndex+3)
		queryBuilder.WriteString(", NOW())")

		args = append(args, event.EventType, eventDataJSON, event.IdempotencyKey, string(ports.OutboxStatusPending))
		argIndex += paramsPerEvent
	}

	queryBuilder.WriteString(" ON CONFLICT (idempotency_key) DO NOTHING RETURNING id, created_at")
	query := queryBuilder.String()

	executor := r.getQueryExecutor()
	var rows *sql.Rows
	var err error
	if r.tx != nil {
		stmt, prepErr := executor.PrepareContext(ctx, query)
		if prepErr != nil {
			return fmt.Errorf("failed to prepare batch insert statement: %w", prepErr)
		}
		defer stmt.Close()
		rows, err = stmt.QueryContext(ctx, args...)
	} else {
		rows, err = executor.QueryContext(ctx, query, args...)
	}
	if err != nil {
		return fmt.Errorf("failed to execute batch insert: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
		}
	}()

	eventIndex := 0
	insertedEvents := make(map[int]bool)

	for rows.Next() {
		if eventIndex >= len(events) {
			return fmt.Errorf("more rows returned than expected: got %d, expected %d", eventIndex+1, len(events))
		}

		err := rows.Scan(&events[eventIndex].ID, &events[eventIndex].CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan batch insert result: %w", err)
		}
		insertedEvents[eventIndex] = true
		eventIndex++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating batch insert results: %w", err)
	}

	if eventIndex == 0 && len(events) > 0 {
		return nil
	}

	return nil
}

func writeInt(b *strings.Builder, n int) {
	b.WriteString(strconv.Itoa(n))
}

func (r *postgresOutboxRepository) saveEventsBatchInChunks(ctx context.Context, events []*ports.OutboxEvent) error {
	for i := 0; i < len(events); i += r.maxBatchSize {
		end := i + r.maxBatchSize
		if end > len(events) {
			end = len(events)
		}

		chunk := events[i:end]
		if err := r.saveEventsBatchSingle(ctx, chunk); err != nil {
			return fmt.Errorf("failed to save events batch chunk [%d:%d]: %w", i, end, err)
		}
	}

	return nil
}
