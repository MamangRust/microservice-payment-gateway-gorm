-- +goose Up
-- +goose StatementBegin
-- Phase 2: operation state machine columns (see transaction equivalent).
ALTER TABLE transfers
    ADD COLUMN IF NOT EXISTS operation_id UUID NOT NULL DEFAULT gen_random_uuid(),
    ADD COLUMN IF NOT EXISTS failure_reason TEXT;

CREATE INDEX IF NOT EXISTS idx_transfers_status_updated_at
    ON transfers (status, updated_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_transfers_status_updated_at;
ALTER TABLE transfers
    DROP COLUMN IF EXISTS failure_reason,
    DROP COLUMN IF EXISTS operation_id;
-- +goose StatementEnd
