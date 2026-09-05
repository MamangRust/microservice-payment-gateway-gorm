-- ClickHouse Schema for Payment Gateway Stats

CREATE TABLE IF NOT EXISTS transaction_events (
    transaction_id UInt64,
    transaction_no String,
    card_number String,
    card_type String,
    card_provider String,
    amount Int64,
    payment_method String,
    merchant_id UInt64,
    merchant_name String,
    status String,
    apikey String,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (merchant_id, created_at);

CREATE TABLE IF NOT EXISTS topup_events (
    topup_id UInt64,
    topup_no String,
    card_number String,
    card_type String,
    card_provider String,
    amount Int64,
    payment_method String,
    status String,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (card_number, created_at);

CREATE TABLE IF NOT EXISTS transfer_events (
    transfer_id UInt64,
    transfer_no String,
    source_card String,
    destination_card String,
    amount Int64,
    status String,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (source_card, created_at);

CREATE TABLE IF NOT EXISTS withdraw_events (
    withdraw_id UInt64,
    withdraw_no String,
    card_number String,
    card_type String,
    amount Int64,
    status String,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (card_number, created_at);

CREATE TABLE IF NOT EXISTS saldo_events (
    card_number String,
    total_balance Int64,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (card_number, created_at);

CREATE TABLE IF NOT EXISTS card_events (
    card_id UInt64,
    user_id UInt64,
    card_number String,
    card_type String,
    card_provider String,
    status String,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (user_id, created_at);

CREATE TABLE IF NOT EXISTS merchant_events (
    merchant_id UInt64,
    user_id UInt64,
    name String,
    email String,
    status String,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (user_id, created_at);
