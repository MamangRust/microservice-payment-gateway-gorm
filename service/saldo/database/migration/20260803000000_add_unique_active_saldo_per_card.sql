-- +goose Up
-- +goose StatementBegin

-- Dedupe active saldos: keep a single row per card (the one with the highest
-- total_balance; on ties the newest). Historically, card creation published a
-- Kafka event (total_balance=0) that raced with the API create (real balance),
-- producing duplicate active rows and non-deterministic balance lookups.
DELETE FROM saldos a
USING saldos b
WHERE a.card_number = b.card_number
  AND a.deleted_at IS NULL
  AND b.deleted_at IS NULL
  AND a.saldo_id <> b.saldo_id
  AND (
    a.total_balance < b.total_balance
    OR (a.total_balance = b.total_balance AND a.saldo_id < b.saldo_id)
  );

-- Enforce one active saldo per card so GetSaldoByCardNumber is deterministic
-- and CreateSaldo can use ON CONFLICT for atomic idempotent upserts.
CREATE UNIQUE INDEX idx_saldos_card_number_active
ON saldos (card_number)
WHERE deleted_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_saldos_card_number_active;

-- +goose StatementEnd
