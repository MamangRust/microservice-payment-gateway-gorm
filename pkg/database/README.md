# Persistence Layer

This package holds the database plumbing shared by every DB-backed service:
connection setup, the bounded-context database names, migration execution and
PostgreSQL error mapping. Access to tables itself lives in each service's
`service/<svc>/repository/`.

## PostgreSQL configuration

PostgreSQL is the primary OLTP store. Connections are opened with **GORM** on top
of the `pgx` stdlib driver (`gorm.io/driver/postgres`).

### One database per bounded context

There are three databases, one per bounded context:

| Database | Context | Services |
|---|---|---|
| `pg_identity` | Identity | `auth`, `role`, `user` |
| `pg_payment` | Payment | `card`, `merchant`, `saldo` |
| `pg_financial` | Financial | `topup`, `transaction`, `transfer`, `withdraw` |

The names and the environment prefixes that select them are declared in
`names.go` (`IdentityDB`/`PaymentDB`/`FinancialDB` and
`IdentityCluster`/`PaymentCluster`/`FinancialCluster`). Each service passes its
cluster prefix to `server.Config.DBCluster`; the mapping is pinned by
`tests/db_contract_test.go`.

## Connecting

Both `NewGormClientWithPrefix` (GORM) and `RunMigrations` (goose) resolve their
connection through the same `resolveClusterConfig` helper (`cluster.go`), so the
connection and the migration can never target different databases.

- `NewGormClient(log)` — opens a connection using the generic `DB` prefix.
- `NewGormClientWithPrefix(log, prefix)` — opens a connection for a cluster
  prefix:

  | Key | Bounded-context prefix (e.g. `DB_PAYMENT`) | Generic `DB` prefix |
  |---|---|---|
  | `<PREFIX>_HOST` / `_PORT` / `_NAME` | **required** — missing value fails startup | falls back to `DB_HOST` / `DB_PORT` / `DB_NAME` |
  | `<PREFIX>_DRIVER` | falls back to `DB_DRIVER` | `DB_DRIVER` |
  | `<PREFIX>_USERNAME` / `_PASSWORD` | falls back to `DB_USERNAME` / `DB_PASSWORD` | `DB_USERNAME` / `DB_PASSWORD` |

A bounded-context prefix is **strict**: if any of `_HOST` / `_PORT` / `_NAME` is
missing, the service fails to start with an error naming the missing key. It does
not fall back to the base keys, so a misconfigured service can never silently
connect to (or migrate) the wrong database. The generic `DB` prefix keeps the
legacy fallback for callers that do not use the cluster model.

`pkg/server` calls `NewGormClientWithPrefix` and `RunMigrations` with
`cfg.DBCluster`, so a service only ever touches its own context's database.

### Connection pool

| Setting | Env (`<PREFIX>_*` → `DB_*`) | Default |
|---|---|---|
| Max open conns | `MAX_OPEN_CONNS` | 100 |
| Max idle conns | `MIN_IDLE_CONNS` | 50 |
| Conn max lifetime | `CONN_MAX_LIFETIME` | 1h |
| Conn max idle time | `CONN_MAX_IDLE_TIME` | 30m |

## Migrations

Schema changes are managed with **goose**, one set of versioned SQL files per
service at `service/<svc>/database/migration/`. They run automatically at service
startup: `pkg/server` calls `RunMigrations(log, prefix, path)` for any service
that sets `MigrationPath`. `RunMigrations` resolves the same `<PREFIX>_*` keys as
the connection, so a service's migrations land in its own context database.

There is no standalone migrate command and no `service/migrate`; startup is the
only migration path. Task running is via `justfile` (see `just --list`).

## Seeding

There is no seeder service or `just seeder` recipe. Test data is created by the
integration suites under `tests/`.

## Error mapping

`errors.go` exposes helpers such as `IsUniqueViolation` (SQLSTATE `23505`) so
repositories can translate PostgreSQL constraint failures into domain errors
(e.g. HTTP 409) instead of leaking driver errors.
