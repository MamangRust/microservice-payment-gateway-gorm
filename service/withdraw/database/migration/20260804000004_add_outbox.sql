-- Phase 3: Transactional outbox for Kafka events.
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS outbox_events (
    id            BIGSERIAL PRIMARY KEY,
    aggregate_type VARCHAR(64)  NOT NULL,
    aggregate_id   VARCHAR(128) NOT NULL,
    event_id       VARCHAR(128) NOT NULL UNIQUE,
    event_type     VARCHAR(128) NOT NULL,
    event_version  INT          NOT NULL DEFAULT 1,
    payload        JSONB        NOT NULL,
    status         VARCHAR(20)  NOT NULL DEFAULT 'pending',
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
