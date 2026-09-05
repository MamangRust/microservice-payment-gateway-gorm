-- +goose Up
-- +goose StatementBegin
-- Phase 2: operation state machine columns.
-- operation_id correlates the financial command with its idempotency record and
-- recovery worker actions; failure_reason records why a command left the happy path.
-- Both columns are backward compatible: existing rows get a generated UUID and a
-- NULL failure reason, and the existing status column already accepts the new
-- values (processing, compensating, compensated, unknown).
ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS operation_id UUID NOT NULL DEFAULT gen_random_uuid(),
    ADD COLUMN IF NOT EXISTS failure_reason TEXT;

CREATE INDEX IF NOT EXISTS idx_transactions_status_updated_at
    ON transactions (status, updated_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_transactions_status_updated_at;
ALTER TABLE transactions
    DROP COLUMN IF EXISTS failure_reason,
    DROP COLUMN IF EXISTS operation_id;
-- +goose StatementEnd
