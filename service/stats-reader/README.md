# Analytics Engine and Real-Time Reporting

This document describes the analytical infrastructure used for sub-second reporting and historical data aggregation across the platform.

## ClickHouse OLAP Cluster

The system utilizes ClickHouse as its primary analytical (OLAP) engine. It is configured for high-throughput ingestion and sub-second query performance.

### Data Ingestion Pipeline

All financial events from the Distributed Event Bus are consumed by the `stats-writer` service and asynchronously ingested into ClickHouse. This ensures that analytical processing has zero performance impact on the primary transactional (OLTP) database.

### Analytical RPC Interface

The `stats-reader` service provides over 125 specialized gRPC procedures for querying platform-wide metrics, including:
- Transaction throughput and success rates.
- Merchant-specific performance data.
- System-wide reconciliation reports.
- User behavioral statistics for the AI Security engine.

## Query Performance and Sharding

ClickHouse utilize a MergeTree engine family for efficient data storage and indexing.
- Distributed Tables: In production, tables are distributed across multiple nodes to facilitate parallel query execution.
- Materialized Views: Critical aggregations are pre-computed using materialized views to ensure sub-millisecond response times for common dashboard metrics.

## Observability of Analytics

The health and performance of the analytics engine are monitored via the Unified OTLP pipeline. Key metrics tracked include:
- Ingestion lag (Kafka offset to ClickHouse commit).
- Query execution latency percentiles (p50, p90, p99).
- Shard synchronization status.
