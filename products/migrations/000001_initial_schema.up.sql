CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS outbox (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(255) NOT NULL,
    event_data JSONB NOT NULL,
    idempotency_key VARCHAR(255) UNIQUE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    published_at TIMESTAMP WITHOUT TIME ZONE,
    retry_count INT DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'published', 'failed', 'dlq')),
    dlq_reason TEXT
);

CREATE INDEX idx_outbox_status_created ON outbox(status, created_at) WHERE status = 'pending';
CREATE INDEX idx_outbox_idempotency_key ON outbox(idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX idx_outbox_status_created_at ON outbox(status, created_at);

