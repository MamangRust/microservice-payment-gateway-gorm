-- +goose Up
-- +goose StatementBegin
-- Phase 4: append-only balance ledger.
-- Existing balances are represented by an opening entry so reconciliation can
-- compare current saldo against opening + net ledger mutations.
CREATE TABLE IF NOT EXISTS balance_ledger (
    entry_id       BIGSERIAL PRIMARY KEY,
    operation_id   VARCHAR(128) NOT NULL,
    card_number    VARCHAR(16)  NOT NULL,
    direction      VARCHAR(16)  NOT NULL,
    amount         BIGINT       NOT NULL CHECK (amount >= 0),
    balance_before BIGINT       NOT NULL,
    balance_after  BIGINT       NOT NULL CHECK (balance_after >= 0),
    source_type    VARCHAR(64)  NOT NULL,
    source_id      VARCHAR(128),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT balance_ledger_direction_check
        CHECK (direction IN ('opening', 'debit', 'credit', 'reversal')),
    CONSTRAINT balance_ledger_operation_card_direction_unique
        UNIQUE (operation_id, card_number, direction)
);

CREATE INDEX IF NOT EXISTS idx_balance_ledger_card_created
    ON balance_ledger (card_number, created_at, entry_id);
CREATE INDEX IF NOT EXISTS idx_balance_ledger_operation
    ON balance_ledger (operation_id);

-- Keep every newly-created saldo reconcilable from the first write. The
-- opening entry is inserted in the same transaction as the saldo INSERT.
CREATE OR REPLACE FUNCTION create_saldo_opening_ledger()
RETURNS trigger AS $$
BEGIN
    INSERT INTO balance_ledger (
        operation_id, card_number, direction, amount,
        balance_before, balance_after, source_type, source_id, created_at
    ) VALUES (
        'saldo-opening:' || NEW.saldo_id::text,
        NEW.card_number,
        'opening',
        NEW.total_balance,
        0,
        NEW.total_balance,
        'saldo_opening',
        NEW.saldo_id::text,
        COALESCE(NEW.created_at, now())
    ) ON CONFLICT (operation_id, card_number, direction) DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS saldos_opening_ledger ON saldos;
CREATE TRIGGER saldos_opening_ledger
    AFTER INSERT ON saldos
    FOR EACH ROW EXECUTE FUNCTION create_saldo_opening_ledger();

-- Ledger entries are immutable: corrections must be represented by a new
-- reversal/compensation entry, never UPDATE or DELETE.
CREATE OR REPLACE FUNCTION prevent_balance_ledger_mutation()
RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'balance_ledger is append-only';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS balance_ledger_immutable ON balance_ledger;
CREATE TRIGGER balance_ledger_immutable
    BEFORE UPDATE OR DELETE ON balance_ledger
    FOR EACH ROW EXECUTE FUNCTION prevent_balance_ledger_mutation();

-- Backfill a deterministic opening entry for existing active balances.
INSERT INTO balance_ledger (
    operation_id, card_number, direction, amount,
    balance_before, balance_after, source_type, source_id, created_at
)
SELECT
    'saldo-opening:' || saldo_id::text,
    card_number,
    'opening',
    total_balance,
    0,
    total_balance,
    'saldo_opening',
    saldo_id::text,
    COALESCE(created_at, now())
FROM saldos
WHERE deleted_at IS NULL
ON CONFLICT (operation_id, card_number, direction) DO NOTHING;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS balance_ledger_immutable ON balance_ledger;
DROP TRIGGER IF EXISTS saldos_opening_ledger ON saldos;
DROP FUNCTION IF EXISTS prevent_balance_ledger_mutation();
DROP FUNCTION IF EXISTS create_saldo_opening_ledger();
DROP TABLE IF EXISTS balance_ledger;
-- +goose StatementEnd
