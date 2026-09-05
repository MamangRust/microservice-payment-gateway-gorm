-- +goose Up
-- +goose StatementBegin
-- Phase 4.1: signed ledger deltas make reversal/adjustment semantics explicit.
ALTER TABLE balance_ledger
    ADD COLUMN IF NOT EXISTS delta BIGINT,
    ADD COLUMN IF NOT EXISTS note TEXT;

-- The append-only trigger (balance_ledger_immutable) blocks UPDATE/DELETE on
-- balance_ledger. The one-time backfill below is an exception, so the trigger
-- is suspended for the duration of this migration and re-enabled afterwards.
ALTER TABLE balance_ledger DISABLE TRIGGER balance_ledger_immutable;

UPDATE balance_ledger
SET delta = CASE
    WHEN direction = 'debit' THEN -amount
    ELSE amount
END
WHERE delta IS NULL;

ALTER TABLE balance_ledger ENABLE TRIGGER balance_ledger_immutable;

ALTER TABLE balance_ledger
    ALTER COLUMN delta SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_balance_ledger_card_delta
    ON balance_ledger (card_number, delta, created_at, entry_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'balance_ledger_delta_direction_check'
    ) THEN
        ALTER TABLE balance_ledger
            ADD CONSTRAINT balance_ledger_delta_direction_check CHECK (
                (direction = 'opening' AND delta >= 0)
                OR (direction = 'debit' AND delta < 0)
                OR (direction = 'credit' AND delta > 0)
                OR (direction = 'reversal' AND delta <> 0)
            );
    END IF;
END $$;

-- Re-create the opening trigger so it writes the signed delta column too.
CREATE OR REPLACE FUNCTION create_saldo_opening_ledger()
RETURNS trigger AS $$
BEGIN
    INSERT INTO balance_ledger (
        operation_id, card_number, direction, amount, delta,
        balance_before, balance_after, source_type, source_id, created_at
    ) VALUES (
        'saldo-opening:' || NEW.saldo_id::text,
        NEW.card_number,
        'opening',
        NEW.total_balance,
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

-- Durable mismatch queue. A mismatch is retained until an operator explicitly
-- resolves it after the balance and ledger agree again.
CREATE TABLE IF NOT EXISTS reconciliation_queue (
    queue_id                BIGSERIAL PRIMARY KEY,
    saldo_id                INTEGER NOT NULL,
    card_number             VARCHAR(16) NOT NULL,
    current_balance         BIGINT NOT NULL,
    ledger_balance          BIGINT NOT NULL,
    difference              BIGINT NOT NULL,
    ledger_entries          BIGINT NOT NULL,
    status                  VARCHAR(20) NOT NULL DEFAULT 'pending',
    mismatch_count          BIGINT NOT NULL DEFAULT 1,
    first_seen_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at             TIMESTAMPTZ,
    resolution_operation_id VARCHAR(128),
    resolution_note         TEXT,
    CONSTRAINT reconciliation_queue_status_check
        CHECK (status IN ('pending', 'investigating', 'resolved'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_reconciliation_queue_active_saldo
    ON reconciliation_queue (saldo_id)
    WHERE status IN ('pending', 'investigating');
CREATE INDEX IF NOT EXISTS idx_reconciliation_queue_status_seen
    ON reconciliation_queue (status, last_seen_at, queue_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_reconciliation_queue_status_seen;
DROP INDEX IF EXISTS idx_reconciliation_queue_active_saldo;
DROP TABLE IF EXISTS reconciliation_queue;
DROP INDEX IF EXISTS idx_balance_ledger_card_delta;
ALTER TABLE balance_ledger
    DROP CONSTRAINT IF EXISTS balance_ledger_delta_direction_check;
ALTER TABLE balance_ledger DROP COLUMN IF EXISTS note;
ALTER TABLE balance_ledger DROP COLUMN IF EXISTS delta;
-- +goose StatementEnd
