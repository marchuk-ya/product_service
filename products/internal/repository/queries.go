package repository

const (
	queryCreateProduct = `
		INSERT INTO products (name, price, created_at)
		VALUES ($1, $2, NOW())
		RETURNING id, created_at
	`

	queryGetProductByID = `
		SELECT id, name, price, created_at
		FROM products
		WHERE id = $1
	`

	queryListProducts = `
		SELECT 
			id, 
			name, 
			price, 
			created_at,
			COUNT(*) OVER() as total
		FROM products
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	queryDeleteProduct = `
		DELETE FROM products WHERE id = $1
	`
)

const (
	querySaveOutboxEvent = `
		INSERT INTO outbox (event_type, event_data, idempotency_key, status, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (idempotency_key) DO NOTHING
		RETURNING id, created_at
	`

	queryGetPendingEvents = `
		SELECT id, event_type, event_data, idempotency_key, created_at, published_at, retry_count, status
		FROM outbox
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`

	queryMarkAsPublished = `
		UPDATE outbox
		SET status = $1, published_at = NOW()
		WHERE id = $2
	`

	queryMarkAsFailed = `
		UPDATE outbox
		SET status = $1, retry_count = $2
		WHERE id = $3
	`

	queryCheckIdempotencyKey = `
		SELECT EXISTS(SELECT 1 FROM outbox WHERE idempotency_key = $1)
	`

	queryMoveToDLQ = `
		UPDATE outbox
		SET status = $1, retry_count = retry_count + 1, dlq_reason = $3
		WHERE id = $2
	`

	querySaveEventsBatch = `
		INSERT INTO outbox (event_type, event_data, idempotency_key, status, created_at)
		VALUES 
	`
)

