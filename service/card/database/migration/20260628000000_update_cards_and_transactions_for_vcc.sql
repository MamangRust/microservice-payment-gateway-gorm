-- +goose Up
-- +goose StatementBegin
ALTER TABLE cards
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS credit_limit INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS outstanding_balance INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reward_points INT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS billing_cycles (
    billing_id SERIAL PRIMARY KEY,
    card_number VARCHAR(16) NOT NULL REFERENCES cards (card_number),
    cycle_start TIMESTAMP NOT NULL,
    cycle_end TIMESTAMP NOT NULL,
    amount_due INT NOT NULL DEFAULT 0,
    due_date TIMESTAMP NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'unpaid',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT billing_cycles_period_valid CHECK (cycle_end > cycle_start),
    CONSTRAINT billing_cycles_amount_valid CHECK (amount_due >= 0),
    CONSTRAINT billing_cycles_due_date_valid CHECK (due_date >= cycle_end)
);

CREATE INDEX IF NOT EXISTS idx_billing_cycles_card_number ON billing_cycles (card_number);
-- A statement period is unique per card. This is the concurrency/idempotency
-- guard for retries and multiple scheduler instances.
CREATE UNIQUE INDEX IF NOT EXISTS uq_billing_cycles_card_period
    ON billing_cycles (card_number, cycle_start);
CREATE INDEX IF NOT EXISTS idx_billing_cycles_status ON billing_cycles (status);
CREATE INDEX IF NOT EXISTS idx_billing_cycles_due_date ON billing_cycles (due_date);
CREATE INDEX IF NOT EXISTS idx_cards_status ON cards (status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_billing_cycles_due_date;
DROP INDEX IF EXISTS idx_billing_cycles_status;
DROP INDEX IF EXISTS uq_billing_cycles_card_period;
DROP INDEX IF EXISTS idx_billing_cycles_card_number;
DROP TABLE IF EXISTS billing_cycles;
ALTER TABLE cards
    DROP COLUMN IF EXISTS reward_points,
    DROP COLUMN IF EXISTS outstanding_balance,
    DROP COLUMN IF EXISTS credit_limit,
    DROP COLUMN IF EXISTS status;
-- +goose StatementEnd
