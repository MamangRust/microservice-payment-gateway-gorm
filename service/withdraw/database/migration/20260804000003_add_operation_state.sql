-- +goose Up
-- +goose StatementBegin
ALTER TABLE withdraws
    ADD COLUMN IF NOT EXISTS operation_id UUID NOT NULL DEFAULT gen_random_uuid(),
    ADD COLUMN IF NOT EXISTS failure_reason TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE withdraws
    DROP COLUMN IF EXISTS failure_reason,
    DROP COLUMN IF EXISTS operation_id;
-- +goose StatementEnd
