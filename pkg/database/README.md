# Persistence Layer and Database Schema

This document outlines the data storage architecture, migration workflows, and seeding strategies for the PostgreSQL primary transactional database.

## PostgreSQL Configuration

The system utilizes PostgreSQL 17-alpine as its primary OLTP store, configured for high-availability using a primary-replica topology.

### Logical Schema Organization

- Authentication: Identity stores, role-based access control (RBAC), and session metadata.
- Core Transactions: Atomic records of all financial events including status, type, and cross-reference identifiers.
- Balances: Real-time ledger accounting for all user and merchant entities.

## Migration Framework

Schema evolutions are managed using the `goose` migration tool. All migrations are version-controlled and idempotent.

To apply migrations locally:

```bash
just migrate
```

Migrations are located in `pkg/database/migrations/`.

## Data Seeding Strategy

For performance testing and development parity, a high-volume seeding engine is implemented.

- Go Seeders: Programmatic data generation located in `pkg/database/seeder/`.
- SQL Seeds: A high-fidelity `seeder.sql` containing a baseline of 100+ users and 1,500+ heterogeneous records.

To seed the database:

```bash
just seeder
```

## Connection Pooling

PgBouncer is utilized for server-side connection pooling to prevent saturation of the PostgreSQL backend during high-concurrency event bursts. All microservices connect to PostgreSQL via the pgbouncer service endpoint.
