# Analytics Engine and Real-Time Reporting

This document describes the analytical infrastructure used for reporting and
historical data aggregation across the platform.

## ClickHouse OLAP Store

ClickHouse is the analytical (OLAP) engine. A single ClickHouse instance holds
all event tables; there are no distributed tables and no materialized views.
Aggregations are computed at query time.

### Data Ingestion Pipeline

Domain events are published to Kafka (`stats-topic-*`) by the transactional
services and consumed by `stats-writer`, which batch-inserts them into
ClickHouse. This keeps analytical processing off the primary transactional
(OLTP) database.

- Batch size: 1000 rows, flushed at most every 5 seconds.
- Consumer group: `stats-writer-group`.
- Deduplication is in-memory only (48-hour window), so aggregates depend on the
  consumer process staying alive across redeliveries.

### Analytical RPC Interface

`stats-reader` exposes **109 gRPC procedures across 22 services** for querying
platform-wide metrics, including:

- Transaction throughput and success rates.
- Merchant-specific performance data.
- Card, topup, transfer, withdraw and saldo statistics.

The reader listens on `:50062` by default (override with
`STATS_READER_LISTEN_ADDR`) and queries ClickHouse directly on every request —
there is no cache in the reader; caching lives in `apigateway` (5-minute TTL).

## Query Performance

All seven tables use the `MergeTree` engine with
`PARTITION BY toYYYYMM(created_at)`, which keeps period-bounded queries cheap.
There is no `FINAL` and no `ReplacingMergeTree`, so queries read rows as
inserted.

## Observability of Analytics

The health and performance of the analytics engine are monitored via the
Unified OTLP pipeline. Key metrics tracked include:

- Ingestion lag (Kafka offset to ClickHouse commit).
- Query execution latency percentiles (p50, p90, p99).
