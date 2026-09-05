-- Phase 3: Transactional outbox for Kafka events (saldo).
-- Saldo balance changes publish a durable saldo.changed event so downstream
-- consumers (stats-writer, cache invalidator) can react even if the request
-- process crashes before the synchronous cache delete.
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS outbox_events (
    id            BIGSERIAL PRIMARY KEY,
    aggregate_type VARCHAR(64)  NOT NULL,         -- e.g. 'saldo'
    aggregate_id   VARCHAR(128) NOT NULL,         -- card number
    event_id       VARCHAR(128) NOT NULL UNIQUE,   -- UUID, dedup key for consumers
    event_type     VARCHAR(128) NOT NULL,          -- e.g. 'saldo.changed'
    event_version  INT          NOT NULL DEFAULT 1,
    payload        JSONB        NOT NULL,
    status         VARCHAR(20)  NOT NULL DEFAULT 'pending', -- pending/processing/published/failed
    attempts       INT          NOT NULL DEFAULT 0,
    last_error     TEXT,
    available_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    published_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outbox_status_available
    ON outbox_events (status, available_at)
    WHERE status IN ('pending', 'failed');
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS outbox_events;
-- +goose StatementEnd
