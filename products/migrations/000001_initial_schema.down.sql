DROP INDEX IF EXISTS idx_outbox_status_created_at;
DROP INDEX IF EXISTS idx_outbox_idempotency_key;
DROP INDEX IF EXISTS idx_outbox_status_created;
DROP TABLE IF EXISTS outbox;
DROP TABLE IF EXISTS products;

