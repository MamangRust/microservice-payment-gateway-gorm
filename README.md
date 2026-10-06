# Distributed Microservices Architecture — Payment Gateway Platform

A production-grade, resilient, and fully observable **payment gateway microservices backend**
written in **Go (Golang)**. Financial workloads — identity, cards, merchants, balances, topups,
transactions, transfers, withdrawals — are split across self-contained, independently deployable
services that communicate synchronously over **gRPC** and asynchronously over **Apache Kafka**,
behind a unified **REST API Gateway** (Echo + NGINX).

Analytical reads never touch the transactional tier: a **CQRS-style OLAP layer** powered by
**ClickHouse** is fed by `stats-writer` (a Kafka consumer) and served by `stats-reader`, a gRPC
analytics service returning monthly/yearly volumes, payment-method breakdowns, and transaction
status distributions.

Persistence is split into **three independent PostgreSQL 17 clusters — `pg_identity`,
`pg_payment`, `pg_financial` — one per bounded context, each fronted by its own PgBouncer
pooler**. A service can only ever reach the cluster its context owns, so cross-context schema
coupling is impossible by construction. A missing cluster prefix fails startup instead of
silently connecting to the wrong database.

---

## Table of Contents

1. [Key Features](#key-features)
2. [Architecture Overview](#architecture-overview)
3. [Service Catalog](#service-catalog)
4. [Database Layer — PostgreSQL Cluster & PgBouncer](#database-layer--postgresql-cluster--pgbouncer)
5. [Internal Service Architecture](#internal-service-architecture)
6. [Data & Event Flow](#data--event-flow)
7. [OLAP Analytics Layer](#olap-analytics-layer)
8. [AI Security Service](#ai-security-service)
9. [Observability Architecture](#observability-architecture)
10. [Deployment Architectures](#deployment-architectures)
11. [Technology Stack](#technology-stack)
12. [Getting Started](#getting-started)
13. [Port Map Registry](#port-map-registry)
14. [Makefile / Justfile Reference](#makefile--justfile-reference)
15. [Workspace Directory Tree](#workspace-directory-tree)

---

## Key Features

| Domain | Capabilities |
| :--- | :--- |
| **Auth & Users** | Registration, login, stateless JWT access/refresh token lifecycle, password reset, OTP email verification, `GetMe` profile resolver |
| **Roles & RBAC** | Permission configuration, granular access control, sub-second permission evaluation cached in Redis; role↔user assignment lives in its own `UserRoleService` (`CreateUserRole` / `DeleteUserRole` / `FindByUserId`) |
| **Cards & VCC** | Virtual and debit card CRUD, soft-delete, activation/suspension toggles, multi-dimensional card analytics (daily/monthly/yearly topup, withdraw, transfer) |
| **Merchants** | Merchant onboarding, profile details, business data, performance reports, full soft-delete & restore |
| **Saldo (Balance)** | High-throughput thread-safe real-time balance calculation with optimistic concurrency locks |
| **Topup** | Balance-loading ledger supporting multiple payment methods, detailed logging, soft-delete audit records |
| **Transaction** | Central financial audit ledger with global search filters, status tracking, monthly/yearly volume reports |
| **Transfer** | Peer-to-peer card-to-card / user-to-user settlement with synchronized debit/credit and event-driven logging |
| **Withdraw** | Settlement from cards to external accounts/banks with daily threshold limits and status pipelines |
| **OLAP Analytics** | ClickHouse warehouse fed by `stats-writer` (Kafka consumer) and served by `stats-reader` (gRPC analytics) |
| **Email Worker** | Kafka-driven worker dispatching OTPs, login alerts, merchant notices, and transfer/topup invoices over SMTP |
| **AI Security** | Python gRPC service consuming Kafka events with a Redis feature store for fraud/anomaly detection |
| **Persistence** | Three PostgreSQL 17 clusters — one per bounded context — each fronted by its own PgBouncer pooler |
| **Observability** | Prometheus + Grafana, Loki + Promtail, Jaeger + OpenTelemetry, Pyroscope profiling, Node/Kafka/Postgres/ClickHouse exporters, Alertmanager |
| **Deployment** | Docker Compose (full stack + infra-only), Kubernetes manifests with HPA, ArgoCD GitOps |

---

## Architecture Overview

Each service is a logical, decoupled Go binary inside `service/` with its own gRPC boundary. The
API Gateway is the only public edge: it authenticates the JWT, then translates REST/JSON into
downstream gRPC calls.

### Core Architecture Principles

- **Domain-Driven Boundary Isolation** — every service owns its database, cache, and logic. Cross-boundary database sharing is forbidden.
- **Three-Cluster Persistence** — one PostgreSQL instance and one PgBouncer pooler per bounded context (`pg_identity`, `pg_payment`, `pg_financial`). GORM clients and Goose migrations resolve their connection from the same per-context prefix; a missing prefix fails startup.
- **Dedicated PgBouncer Pooling** — each cluster is fronted by its own pooler so concurrent services cannot exhaust PostgreSQL sockets.
- **Clean Architecture** — `handler → service → repository`, wired in the service bootstrap.
- **OLAP & CQRS** — OLTP writes go to PostgreSQL; analytical reads go to ClickHouse via `stats-reader`.
- **Event-Driven Resilience** — Kafka decouples email delivery and stats materialization from the transactional path.
- **OTel Telemetry Integration** — trace IDs propagate from the REST edge through gRPC down to PostgreSQL, ClickHouse, and Redis.

```mermaid
graph TB
    classDef client fill:#0f172a,stroke:#38bdf8,color:#e0f2fe,stroke-width:2px,font-weight:bold
    classDef gateway fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px,font-weight:bold
    classDef domain fill:#1e1b4b,stroke:#818cf8,color:#e0e7ff,stroke-width:1.5px
    classDef infra fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef obs fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1.5px
    classDef event fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px
    classDef olap fill:#1e293b,stroke:#a855f7,color:#f3e8ff,stroke-width:1.5px

    Client["Client Applications<br/>Web / Mobile / API"]:::client

    subgraph APIGateway["API Gateway — NGINX + Echo"]
        direction LR
        REST["REST API Endpoints<br/>/api/*"]
        Swagger["Swagger UI<br/>/swagger/index.html"]
        AuthMW["JWT Auth<br/>Middleware"]
    end
    class APIGateway gateway

    Client --> APIGateway

    subgraph BusinessServices["Business Domain Services"]
        direction TB

        subgraph IdentityDomain["Identity & Access"]
            AUTH["Auth Service<br/>JWT & OTP verification"]
            USER["User Service<br/>Profile management"]
            ROLE["Role Service<br/>RBAC + UserRole RPC"]
        end

        subgraph MerchantDomain["Merchant"]
            MERCH["Merchant Service<br/>Onboarding & profiling"]
        end

        subgraph FinanceDomain["Finance & Ledger"]
            CARD["Card Service<br/>VCC & card analytics"]
            SALDO["Saldo Service<br/>Real-time balance"]
        end

        subgraph MovementDomain["Fund Movements"]
            TOPUP["Topup Service<br/>Balance funding"]
            TXN["Transaction Service<br/>Central audit register"]
            TRANSFER["Transfer Service<br/>P2P settlement"]
            WITHDRAW["Withdraw Service<br/>Outbound settlement"]
        end
    end
    class BusinessServices domain

    subgraph OLAPEngine["OLAP & Analytics Layer"]
        direction TB
        WRITER["Stats Writer<br/>Kafka consumer"]:::olap
        READER["Stats Reader<br/>gRPC query service :50062"]:::olap
        CLICKHOUSE[("ClickHouse OLAP<br/>Analytics DB")]:::infra
    end

    APIGateway -->|"gRPC — OLTP"| BusinessServices
    APIGateway -->|"gRPC — OLAP"| READER

    subgraph Persistence["PostgreSQL Cluster Tier — 3 contexts"]
        direction LR
        subgraph IdentityPG["Identity"]
            PGB_ID["PgBouncer :6432"]:::infra
            PG_ID[("pg_identity")]:::infra
        end
        subgraph PaymentPG["Payment"]
            PGB_PAY["PgBouncer :6433"]:::infra
            PG_PAY[("pg_payment")]:::infra
        end
        subgraph FinancialPG["Financial"]
            PGB_FIN["PgBouncer :6434"]:::infra
            PG_FIN[("pg_financial")]:::infra
        end
    end

    PGB_ID --> PG_ID
    PGB_PAY --> PG_PAY
    PGB_FIN --> PG_FIN

    BusinessServices -->|"SQL via per-context PgBouncer"| Persistence

    REDIS[("Redis Cluster<br/>6 nodes")]:::infra
    KAFKA[("Kafka KRaft<br/>Event bus")]:::event
    PYRO["Pyroscope<br/>Continuous profiler"]:::obs

    BusinessServices -->|"Cache / Invalidate"| REDIS
    BusinessServices -->|"Publish events"| KAFKA
    BusinessServices -.->|"Profiles"| PYRO

    subgraph EventConsumers["Event-Driven Consumers"]
        EMAIL["Email Service<br/>SMTP worker"]:::event
        AIS["AI Security<br/>Python fraud detector"]:::event
    end

    KAFKA -->|"Consume"| EMAIL
    KAFKA -->|"Consume"| WRITER
    KAFKA -->|"Consume"| AIS
    AIS --> REDIS
    WRITER -->|"Batch insert"| CLICKHOUSE
    READER -->|"Aggregate queries"| CLICKHOUSE
    READER -->|"Cache stats"| REDIS

    subgraph Observability["Observability Stack"]
        direction LR
        PROM["Prometheus"]
        LOKI["Loki"]
        JAEGER["Jaeger"]
        GRAFANA["Grafana"]
        OTEL["OTel Collector"]
        PROMTAIL["Promtail"]
        NODEX["Node Exporter"]
        KAFKAX["Kafka Exporter"]
        PGX["Postgres Exporter"]
        CHX["ClickHouse Exporter"]
        ALERTMGR["Alertmanager"]
    end
    class Observability obs

    BusinessServices -.->|"/metrics"| PROM
    BusinessServices -.->|"OTLP"| OTEL
    READER -.->|"/metrics"| PROM
    WRITER -.->|"/metrics"| PROM
    OTEL -.-> JAEGER
    PROMTAIL -.-> LOKI
    NODEX -.-> PROM
    KAFKAX -.-> PROM
    PGX -.-> PROM
    CHX -.-> PROM
    PROM -.-> ALERTMGR
    PROM -.-> GRAFANA
    LOKI -.-> GRAFANA
    JAEGER -.-> GRAFANA
```

---

## Service Catalog

The workspace ships **12 domain services**, one REST API gateway, two OLAP workers, one AI
security worker, and an email worker.

| # | Service | Bounded Context | gRPC | Responsibility |
|---|---------|-----------------|------|----------------|
| 1 | `apigateway` | — | — (REST `:5000`) | REST/JSON edge, Swagger UI, JWT middleware, gRPC fan-out |
| 2 | `auth` | identity | `50051` | Register, login, refresh, password reset, OTP |
| 3 | `role` | identity | `50052` | Role CRUD + `UserRoleService` (create / delete / find-by-user) |
| 4 | `card` | payment | `50053` | Debit & virtual card management, card analytics |
| 5 | `merchant` | payment | `50054` | Merchant onboarding and profiling |
| 6 | `user` | identity | `50055` | User profiles, soft-delete/restore |
| 7 | `saldo` | payment | `50056` | Real-time balance with optimistic locking |
| 8 | `topup` | financial | `50057` | Balance funding ledger |
| 9 | `transaction` | financial | `50058` | Central audit register |
| 10 | `transfer` | financial | `50059` | P2P card-to-card transfer |
| 11 | `withdraw` | financial | `50060` | Outbound fund settlement |
| 12 | `email` | — | — | Kafka consumer → SMTP notifications |
| 13 | `stats-writer` | ClickHouse | — | Kafka consumer → ClickHouse |
| 14 | `stats-reader` | ClickHouse | `50062` | gRPC analytics over ClickHouse |
| 15 | `ai-security` | — | `50051` | Python Kafka consumer + Redis feature store for fraud detection |

```mermaid
graph LR
    classDef svc fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1px
    classDef gw fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px,font-weight:bold
    classDef support fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1px
    classDef olap fill:#1e293b,stroke:#a855f7,color:#f3e8ff,stroke-width:1px

    API["API Gateway<br/>Echo + REST + Swagger"]:::gw

    subgraph Identity["Identity (3)"]
        A1["auth"]:::svc
        A2["user"]:::svc
        A3["role"]:::svc
    end

    subgraph Payment["Payment (3)"]
        P1["card"]:::svc
        P2["merchant"]:::svc
        P3["saldo"]:::svc
    end

    subgraph Financial["Financial (4)"]
        F1["topup"]:::svc
        F2["transaction"]:::svc
        F3["transfer"]:::svc
        F4["withdraw"]:::svc
    end

    subgraph OLAP["OLAP (2)"]
        O1["stats-writer"]:::olap
        O2["stats-reader"]:::olap
    end

    subgraph Support["Workers (2)"]
        S1["email"]:::support
        S2["ai-security"]:::support
    end

    API --> Identity
    API --> Payment
    API --> Financial
    API --> OLAP
```

---

## Database Layer — PostgreSQL Cluster & PgBouncer

The gateway runs **three independent PostgreSQL 17 clusters** — one per bounded context.
"Cluster" means *one dedicated PostgreSQL instance per context*, not a replicated
primary/replica pair. The split keeps identity data, card/merchant/balance data, and the
financial movement ledger physically separate, so a runaway query or a bad migration in one
context cannot take down the others.

**Every cluster is fronted by its own PgBouncer pooler.** Services dial the pooler, never
PostgreSQL directly.

### Topology

| Bounded Context | PostgreSQL instance | Database | PgBouncer (host port) | Owning services |
| :--- | :--- | :--- | :--- | :--- |
| **Identity** | `postgres_identity` | `pg_identity` | `6432` | `auth`, `role`, `user` |
| **Payment** | `postgres_payment` | `pg_payment` | `6433` | `card`, `merchant`, `saldo` |
| **Financial** | `postgres_financial` | `pg_financial` | `6434` | `topup`, `transaction`, `transfer`, `withdraw` |

> **Pool mode differs per environment** — worth knowing when debugging prepared-statement or
> session-state behaviour:
> - **Local (Docker Compose)**: `POOL_MODE=transaction`
> - **Kubernetes**: `POOL_MODE=session`, with `MAX_CLIENT_CONN=1000`, `DEFAULT_POOL_SIZE=20`,
>   `MAX_PREPARED_STATEMENTS=100`
>
> Auth is `scram-sha-256` in both.

```mermaid
graph TB
    classDef svc fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1px
    classDef pool fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px,font-weight:bold
    classDef pg fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef obs fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1px

    subgraph IdentityCtx["Identity Context"]
        direction TB
        S_ID["auth · role · user"]:::svc
        PGB_ID["pgbouncer_identity :6432<br/>transaction pool locally · session in K8s<br/>scram-sha-256"]:::pool
        PG_ID[("postgres_identity<br/>pg_identity")]:::pg
        PX_ID["postgres-exporter"]:::obs
        S_ID -->|"DB_IDENTITY_HOST/PORT/NAME"| PGB_ID
        PGB_ID -->|"bounded server pool"| PG_ID
        PG_ID -.-> PX_ID
    end

    subgraph PaymentCtx["Payment Context"]
        direction TB
        S_PAY["card · merchant · saldo"]:::svc
        PGB_PAY["pgbouncer_payment :6433"]:::pool
        PG_PAY[("postgres_payment<br/>pg_payment")]:::pg
        PX_PAY["postgres-exporter"]:::obs
        S_PAY -->|"DB_PAYMENT_*"| PGB_PAY
        PGB_PAY --> PG_PAY
        PG_PAY -.-> PX_PAY
    end

    subgraph FinancialCtx["Financial Context"]
        direction TB
        S_FIN["topup · transaction<br/>transfer · withdraw"]:::svc
        PGB_FIN["pgbouncer_financial :6434"]:::pool
        PG_FIN[("postgres_financial<br/>pg_financial")]:::pg
        PX_FIN["postgres-exporter"]:::obs
        S_FIN -->|"DB_FINANCIAL_*"| PGB_FIN
        PGB_FIN --> PG_FIN
        PG_FIN -.-> PX_FIN
    end
```

### Connection Lifecycle

```mermaid
sequenceDiagram
    autonumber
    participant SVC as Domain Service<br/>GORM client
    participant PGB as PgBouncer<br/>per-context pooler
    participant PG as PostgreSQL<br/>per-context instance
    participant PX as postgres-exporter

    SVC->>SVC: Resolve DB_<CONTEXT>_* prefix at startup
    SVC->>PGB: Dial pooler host:port<br/>scram-sha-256 auth
    PGB->>PGB: Admit client within max_client_conn
    alt Server slot available
        PGB->>PG: Assign pooled server connection
    else Pool saturated
        PGB-->>SVC: Queue until a slot frees
    end
    SVC->>PGB: BEGIN / SELECT / INSERT / COMMIT
    PGB->>PG: Forward statement on assigned connection
    PG-->>PGB: Result set
    PGB-->>SVC: Rows
    SVC->>PGB: Close client connection
    PGB->>PG: Return server connection to the pool
    PX-->>PX: Scrape pg_stat_database / pg_stat_activity
```

### Environment Contract

There is no generic `DB_HOST` / `DB_PORT` / `DB_NAME`. Each service declares its cluster once in
`service/<name>/cmd/main.go`:

```go
server.Config{
    DBCluster: database.IdentityCluster, // → "DB_IDENTITY"
    RedisCluster: "REDIS_1",
    // ...
}
```

`pkg/database/names.go` is the single source of truth:

| Constant | Env prefix | Database |
| :--- | :--- | :--- |
| `database.IdentityCluster` | `DB_IDENTITY` | `pg_identity` |
| `database.PaymentCluster` | `DB_PAYMENT` | `pg_payment` |
| `database.FinancialCluster` | `DB_FINANCIAL` | `pg_financial` |

Each prefix resolves `<PREFIX>_HOST` / `<PREFIX>_PORT` / `<PREFIX>_NAME`, pointing at the
context's **PgBouncer** service. Because the base `DB_*` keys are deliberately undefined, a
prefix typo produces a startup failure rather than a silent connection to the wrong database —
an important safety property for a financial ledger.

### Migrations

Migrations are **not** a separate step in this repo: each service applies its own **Goose**
migrations at startup against its bounded-context database, selected by the same cluster prefix.
Prefix-aware migration routing means `topup` can never accidentally migrate `pg_identity`.

```mermaid
flowchart LR
    classDef svc fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1.5px
    classDef pool fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px
    classDef pg fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px

    subgraph IdentitySvc["DB_IDENTITY"]
        A["auth"]:::svc
        R["role"]:::svc
        U["user"]:::svc
    end
    subgraph PaymentSvc["DB_PAYMENT"]
        C["card"]:::svc
        M["merchant"]:::svc
        S["saldo"]:::svc
    end
    subgraph FinancialSvc["DB_FINANCIAL"]
        T["topup"]:::svc
        X["transaction"]:::svc
        TR["transfer"]:::svc
        W["withdraw"]:::svc
    end

    A -->|"Goose at startup"| P1["pgbouncer_identity"]:::pool
    R --> P1
    U --> P1
    C -->|"Goose at startup"| P2["pgbouncer_payment"]:::pool
    M --> P2
    S --> P2
    T -->|"Goose at startup"| P3["pgbouncer_financial"]:::pool
    X --> P3
    TR --> P3
    W --> P3

    P1 --> D1[("pg_identity")]:::pg
    P2 --> D2[("pg_payment")]:::pg
    P3 --> D3[("pg_financial")]:::pg
```

### Kubernetes Topology

In-cluster each context becomes a **StatefulSet + PVC** with a **Deployment** pooler in front:

```mermaid
flowchart TB
    classDef pool fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px
    classDef pg fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef svc fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1.5px

    subgraph NS["namespace: payment-gateway"]
        subgraph IdentityK8s["Identity"]
            PK_ID["Deployment pgbouncer-identity<br/>POOL_MODE=session · pool 20 · clients 1000"]:::pool
            PS_ID[("StatefulSet postgres-identity<br/>pg_identity + PVC")]:::pg
        end
        subgraph PaymentK8s["Payment"]
            PK_PAY["Deployment pgbouncer-payment"]:::pool
            PS_PAY[("StatefulSet postgres-payment<br/>pg_payment + PVC")]:::pg
        end
        subgraph FinancialK8s["Financial"]
            PK_FIN["Deployment pgbouncer-financial"]:::pool
            PS_FIN[("StatefulSet postgres-financial<br/>pg_financial + PVC")]:::pg
        end

        SVC_ID["auth · role · user"]:::svc
        SVC_PAY["card · merchant · saldo"]:::svc
        SVC_FIN["topup · transaction · transfer · withdraw"]:::svc
    end

    SVC_ID --> PK_ID
    SVC_PAY --> PK_PAY
    SVC_FIN --> PK_FIN

    PK_ID --> PS_ID
    PK_PAY --> PS_PAY
    PK_FIN --> PS_FIN
```

> PostgreSQL ports are not published to the host in either target — connect through PgBouncer
> (`6432`–`6434`).

---

## Internal Service Architecture

```mermaid
graph TB
    classDef handler fill:#1e3a5f,stroke:#7dd3fc,color:#e0f2fe,stroke-width:1.5px
    classDef service fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1.5px
    classDef repo fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef infra fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1.5px
    classDef shared fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px

    subgraph Service["service/<name>/"]
        direction TB

        CMD["cmd/main.go<br/>Entry point + DBCluster/RedisCluster selection"]

        subgraph Internal["internal wiring"]
            direction TB
            APPS["app/client.go or server.go<br/>Dependency injection"]:::handler
            HANDLER["handler/<br/>gRPC handlers"]:::handler
            MW["middleware/<br/>Interceptors"]:::handler
            SVC["service/<br/>Business logic"]:::service
            CACHE["cache/<br/>Redis cache layer"]:::service
            REPO["repository/<br/>Data access — GORM"]:::repo
        end

        CMD --> APPS
        APPS --> HANDLER
        APPS --> SVC
        APPS --> CACHE
        APPS --> REPO
        HANDLER --> SVC
        SVC --> REPO
        SVC --> CACHE
    end

    subgraph SharedLibs["shared/ — Shared Libraries"]
        direction LR
        DOMAIN["domain/<br/>record / request / response"]:::shared
        OBS["observability/<br/>cache & tracing metrics"]:::shared
        CACHESHARED["cache/<br/>redis_cache.go"]:::shared
        MAPPER["mapper/<br/>Domain ↔ Proto"]:::shared
        ERRORS["errors/ + errorhandler/"]:::shared
    end

    subgraph PkgLibs["pkg/ — Platform Libraries"]
        direction LR
        PKGAUTH["auth/<br/>JWT manager"]:::infra
        PKGKAFKA["kafka/<br/>Producer / consumer"]:::infra
        PKGOTEL["otel/<br/>Tracing + metrics init"]:::infra
        PKGRES["resilience/<br/>Circuit breaker, rate limiter,<br/>load monitor, DependencyGuard"]:::infra
        PKGLOG["logger/<br/>Zap structured logging"]:::infra
        PKGSRV["server/<br/>gRPC bootstrap"]:::infra
        PKGDB["database/<br/>Prefix-aware GORM + Goose + error mapping"]:::infra
        PKGCH["clickhouse/<br/>OLAP connection factory"]:::infra
        PKGREDIS["redis/<br/>Cluster-aware Redis client"]:::infra
        PKGADAPTER["adapter/role · adapter/user_role<br/>guarded gRPC clients"]:::infra
    end

    PB["pb/<br/>Generated protobuf Go code"]:::shared
    PGB_EXT["PgBouncer per context"]:::infra

    REPO --> DOMAIN
    REPO --> PGB_EXT
    SVC --> DOMAIN
    SVC --> OBS
    HANDLER --> PB
    HANDLER --> MAPPER
    APPS --> PKGSRV
    APPS --> PKGOTEL
    APPS --> CACHESHARED
    APPS --> PKGADAPTER
    APPS --> OBS
```

### Cross-Context Access: the Adapter Pattern

`service/auth` and `service/user` need role data but must not open a second database connection.
They use typed gRPC adapters, each wrapped in a `DependencyGuard` (per-call timeout + circuit
breaker + bulkhead):

| Adapter | Package | Backing RPC | Used by |
| :--- | :--- | :--- | :--- |
| `RoleAdapter` | `pkg/adapter/role` | `RoleService` — `FindById`, `FindByName` | `service/auth`, `service/user` |
| `UserRoleAdapter` | `pkg/adapter/user_role` | `UserRoleService` — `CreateUserRole`, `DeleteUserRole` | `service/auth` |

---

## Data & Event Flow

### Synchronous Flow — REST → gRPC → Redis → PgBouncer → PostgreSQL

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant GW as API Gateway<br/>Echo + REST
    participant SVC as Domain Service<br/>gRPC server
    participant REDIS as Redis Cluster
    participant PGB as PgBouncer pooler
    participant DB as Context PostgreSQL

    C->>GW: REST HTTP request GET/POST/PUT/DELETE
    GW->>GW: JWT authentication check
    GW->>SVC: gRPC call with protobuf payload
    SVC->>REDIS: Check cache
    alt Cache hit
        REDIS-->>SVC: Cached response
    else Cache miss
        SVC->>PGB: Acquire pooled connection
        PGB->>DB: Execute SQL via GORM
        DB-->>PGB: Result set
        PGB-->>SVC: Rows
        SVC->>REDIS: Populate cache for next read
    end
    SVC-->>GW: gRPC response payload
    GW-->>C: REST HTTP response JSON
```

### Asynchronous Flow — Kafka Notification & Stats Pipeline

```mermaid
sequenceDiagram
    autonumber
    participant SVC as Transaction / Topup / Transfer
    participant K as Kafka Broker
    participant EMAIL as Email Worker
    participant SMTP as SMTP Server
    participant WRITER as Stats Writer
    participant CH as ClickHouse
    participant AIS as AI Security

    SVC->>K: Publish event transfer.created / topup.success
    K-->>EMAIL: Deliver topic payload
    EMAIL->>EMAIL: Map payload details
    EMAIL->>SMTP: Send styled notification
    SMTP-->>EMAIL: Delivery confirmation
    K-->>WRITER: Deliver stats event
    WRITER->>CH: Batch insert into analytics tables
    CH-->>WRITER: Batch flushed
    K-->>AIS: Deliver event for scoring
    AIS->>AIS: Score against Redis feature store
```

### Fund Movement Flow — Transfer

```mermaid
sequenceDiagram
    autonumber
    participant GW as API Gateway
    participant TR as transfer service
    participant SA as saldo service
    participant CA as card service
    participant PGB as PgBouncer financial / payment
    participant K as Kafka

    GW->>TR: POST /api/transfer
    TR->>CA: Validate source & destination cards
    TR->>SA: Debit source balance (optimistic lock)
    SA->>PGB: UPDATE saldo WHERE version = ?
    PGB-->>SA: 1 row affected
    TR->>SA: Credit destination balance
    SA->>PGB: UPDATE saldo
    TR->>PGB: INSERT transfer + transaction ledger rows
    TR->>K: Publish transfer.created
    TR-->>GW: Transfer response
```

---

## OLAP Analytics Layer

Transactional writes land in the per-context PostgreSQL clusters; analytical reads are served by
ClickHouse through `stats-reader` on port `50062`, which the API Gateway calls with cache-aside
Redis.

```mermaid
graph LR
    classDef olap fill:#1e293b,stroke:#a855f7,color:#f3e8ff,stroke-width:1.5px
    classDef store fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef api fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px,font-weight:bold
    classDef bus fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px

    TXN["topup · transaction<br/>transfer · withdraw · card"]:::olap
    KAFKA[("Kafka<br/>domain events")]:::bus
    WRITER["stats-writer<br/>Kafka consumer"]:::olap
    CH[("ClickHouse<br/>columnar warehouse")]:::store
    READER["stats-reader :50062<br/>gRPC analytics"]:::olap
    GW["API Gateway<br/>/api/card/stats/* · /api/saldo/stats/*<br/>/api/topup/stats/* · /api/transaction/stats/*<br/>/api/transfer/stats/* · /api/withdraw/stats/*"]:::api
    REDIS[("Redis<br/>stats cache")]:::store

    TXN -->|"publish"| KAFKA
    KAFKA -->|"consume"| WRITER
    WRITER -->|"batch insert"| CH
    GW -->|"gRPC"| READER
    READER -->|"aggregate"| CH
    READER -->|"cache-aside"| REDIS
```

---

## AI Security Service

`service/ai-security` is a **Python** worker that consumes the same Kafka event stream and scores
transactions for fraud/anomaly signals using a Redis-backed feature store.

```mermaid
flowchart LR
    classDef bus fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px
    classDef ai fill:#1e293b,stroke:#a855f7,color:#f3e8ff,stroke-width:1.5px
    classDef store fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px

    K[("Kafka<br/>transaction events")]:::bus
    CONS["kafka_consumer.py"]:::ai
    FS["feature_store.py<br/>Redis-backed features"]:::ai
    DET["detector.py<br/>anomaly scoring"]:::ai
    SVC["service.py<br/>gRPC :50051"]:::ai
    R[("Redis Cluster")]:::store

    K --> CONS
    CONS --> FS
    FS --> R
    FS --> DET
    DET --> SVC
```

---

## Observability Architecture

```mermaid
graph TB
    classDef service fill:#1e1b4b,stroke:#818cf8,color:#e0e7ff,stroke-width:1.5px
    classDef collector fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef storage fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1.5px
    classDef viz fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:2px,font-weight:bold

    subgraph Sources["Telemetry Sources"]
        direction TB
        SVCS["12 domain services<br/>+ apigateway + stats-reader/writer + ai-security"]:::service
        KAFKA_SRC["Kafka broker"]:::service
        PG_SRC["3 PostgreSQL clusters<br/>+ 3 PgBouncer poolers"]:::service
        CH_SRC["ClickHouse"]:::service
        NODES["Host / node"]:::service
    end

    subgraph Collectors["Collection Layer"]
        direction TB
        PROM["Prometheus<br/>scrapes /metrics"]:::collector
        PROMTAIL["Promtail<br/>ships container logs"]:::collector
        OTEL["OTel Collector<br/>receives OTLP spans"]:::collector
        NODEX["Node Exporter"]:::collector
        KAFKAX["Kafka Exporter"]:::collector
        PGX["Postgres Exporter"]:::collector
        CHX["ClickHouse Exporter"]:::collector
        PYRO["Pyroscope<br/>continuous profiling"]:::collector
    end

    subgraph Storage["Storage Layer"]
        direction TB
        PROM_TSDB["Prometheus TSDB"]:::storage
        LOKI_STORE["Loki chunks"]:::storage
        JAEGER_STORE["Jaeger trace store"]:::storage
        PYRO_STORE["Pyroscope profiles"]:::storage
    end

    subgraph Visualization["Visualization & Alerting"]
        GRAFANA["Grafana<br/>unified dashboards"]:::viz
        ALERTMGR["Alertmanager<br/>alert routing"]:::viz
    end

    SVCS -->|"/metrics"| PROM
    SVCS -->|"OTLP gRPC"| OTEL
    SVCS -->|"profiles"| PYRO
    SVCS -->|"stdout / stderr"| PROMTAIL
    NODES --> NODEX
    KAFKA_SRC --> KAFKAX
    PG_SRC --> PGX
    CH_SRC --> CHX

    NODEX --> PROM
    KAFKAX --> PROM
    PGX --> PROM
    CHX --> PROM
    PROM --> PROM_TSDB
    PROMTAIL --> LOKI_STORE
    OTEL --> JAEGER_STORE
    PYRO --> PYRO_STORE

    PROM_TSDB --> GRAFANA
    LOKI_STORE --> GRAFANA
    JAEGER_STORE --> GRAFANA
    PYRO_STORE --> GRAFANA
    PROM_TSDB --> ALERTMGR
```

| Pillar | Tooling | What you get |
| :--- | :--- | :--- |
| **Metrics** | Prometheus + Grafana | CPU/memory, request error rates, gRPC latencies, database connection states |
| **Logging** | Loki + Promtail | Structured JSON indexed per service, queryable with LogQL |
| **Tracing** | OpenTelemetry + Jaeger | End-to-end traces across REST → gRPC → PostgreSQL / ClickHouse / Redis |
| **Profiling** | Pyroscope | Continuous CPU/memory profiling to catch allocation leaks in transaction loops |
| **Alerting** | Alertmanager | Notifications on latency spikes and service disconnects |

---

## Deployment Architectures

### Docker Compose (Local Development)

Two compose files live under `deployments/local/`:

- **`docker-compose.yml`** — the whole platform: 3 PostgreSQL + 3 PgBouncer, a 6-node Redis
  cluster, ClickHouse, Kafka, Pyroscope, all Go services, `ai-security`, and the observability
  stack.
- **`docker-compose.infra.yml`** — **infra-only** for native development: the 3
  PostgreSQL/PgBouncer pairs, Redis, Kafka, ClickHouse, and observability, publishing PgBouncer on
  host ports `6432`–`6434`.

```mermaid
flowchart TD
    classDef gateway fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px,font-weight:bold
    classDef core fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1.5px
    classDef infra fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef obs fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1.5px
    classDef event fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px
    classDef olap fill:#1e293b,stroke:#a855f7,color:#f3e8ff,stroke-width:1.5px

    subgraph DockerCompose["docker-compose.yml — Local Environment"]

        subgraph Gateway["Edge"]
            NGINX["NGINX Proxy :80"]
            APIGW["API Gateway :5000<br/>Echo + REST + Swagger"]:::gateway
        end

        subgraph Services["Domain Service Containers"]
            direction TB
            subgraph IdSvc["Identity"]
                AUTH["auth-service"]
                ROLE["role-service"]
                USER["user-service"]
            end
            subgraph PaySvc["Payment"]
                CARD["card-service"]
                MERCH["merchant-service"]
                SALDO["saldo-service"]
            end
            subgraph FinSvc["Financial"]
                TOPUP["topup-service"]
                TXN["transaction-service"]
                TRANSFER["transfer-service"]
                WITHDRAW["withdraw-service"]
            end
            subgraph OLAPSuite["OLAP"]
                SWRITER["stats-writer"]:::olap
                SREADER["stats-reader :50062"]:::olap
            end
        end
        class Services core

        subgraph Infra["Infrastructure Suite"]
            direction TB
            subgraph PGTier["3 × PostgreSQL 17 — each behind its own PgBouncer"]
                PG1[("identity :6432")]
                PG2[("payment :6433")]
                PG3[("financial :6434")]
            end
            subgraph RedisTier["Redis Cluster — 6 nodes"]
                R1[("redis-node-1")]
                R2[("redis-node-2")]
                R3[("redis-node-3")]
                R4[("redis-node-4")]
                R5[("redis-node-5")]
                R6[("redis-node-6")]
            end
            KAFKA[("Kafka :9092")]:::event
            CLICKHOUSE[("ClickHouse :9000/:8123")]:::infra
            PYRO[("Pyroscope :4040")]:::obs
        end

        subgraph Obs["Observability"]
            PROM["Prometheus :9090"]
            GRAFANA["Grafana :3000"]
            LOKI["Loki :3100"]
            JAEGER["Jaeger :16686"]
            OTEL["OTel Collector :4317"]
            NODEX["Node Exporter"]
            KAFKAX["Kafka Exporter"]
            PGX["Postgres Exporter"]
            PROMTAIL["Promtail"]
            ALERTMGR["Alertmanager :9093"]
        end
        class Obs obs

        subgraph Events["Event Consumers"]
            EMAIL["Email Worker"]:::event
            AIS["ai-security :50051"]:::event
        end
    end

    NGINX --> APIGW
    APIGW -->|"gRPC"| Services
    APIGW -->|"gRPC"| SREADER
    Services -->|"SQL via PgBouncer"| PGTier
    Services --> RedisTier
    Services --> KAFKA
    KAFKA --> EMAIL
    KAFKA --> SWRITER
    KAFKA --> AIS
    AIS --> RedisTier
    SWRITER --> CLICKHOUSE
    SREADER --> CLICKHOUSE

    Services -.->|"/metrics"| PROM
    Services -.->|"OTLP"| OTEL
    Services -.->|"profiles"| PYRO
    OTEL -.-> JAEGER
    PROMTAIL -.-> LOKI
    PROM -.-> GRAFANA
    PROM -.-> ALERTMGR
    LOKI -.-> GRAFANA
```

### Kubernetes (Production) + ArgoCD GitOps

Production runs in the `payment-gateway` namespace: one Deployment + Service + HPA per domain
service, one StatefulSet + PVC per PostgreSQL context, one Deployment per PgBouncer pooler, and a
ClickHouse deployment whose schema ships as a ConfigMap. Delivery is GitOps-driven via ArgoCD
(`deployments/gitops/argocd/`) with overlays under `deployments/kubernetes/overlays/`.

```mermaid
flowchart TD
    classDef k8s fill:#0c1222,stroke:#38bdf8,color:#e0f2fe,stroke-width:2px,font-weight:bold
    classDef pod fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1.5px
    classDef hpa fill:#3b0764,stroke:#c084fc,color:#f3e8ff,stroke-width:1px,font-style:italic
    classDef infra fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef obs fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1.5px
    classDef job fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px
    classDef olap fill:#1e293b,stroke:#a855f7,color:#f3e8ff,stroke-width:1.5px

    ARGO["ArgoCD<br/>GitOps controller"]:::k8s
    REPO[("Git repository<br/>deployments/kubernetes")]:::k8s

    subgraph K8S["Kubernetes Cluster — namespace: payment-gateway"]

        subgraph Ingress["Ingress"]
            NGINX["NGINX Ingress + TLS"]:::k8s
            APIGW["API Gateway Pod"]:::pod
        end

        subgraph CorePods["Domain Service Pods + HPAs"]
            direction TB
            subgraph IdentityPods["Identity"]
                AUTH["auth-pod"]:::pod
                USER["user-pod"]:::pod
                ROLE["role-pod"]:::pod
                AUTH_HPA["↕ HPA"]:::hpa
                USER_HPA["↕ HPA"]:::hpa
            end
            subgraph PaymentPods["Payment"]
                CARD["card-pod"]:::pod
                MERCH["merchant-pod"]:::pod
                SALDO["saldo-pod"]:::pod
            end
            subgraph FinancialPods["Financial"]
                TOPUP["topup-pod"]:::pod
                TXN["transaction-pod"]:::pod
                TRANSFER["transfer-pod"]:::pod
                WITHDRAW["withdraw-pod"]:::pod
            end
            subgraph OLAPPods["OLAP"]
                SWRITER["stats-writer-pod"]:::olap
                SREADER["stats-reader-pod"]:::olap
            end
        end

        subgraph DataPods["PostgreSQL Clusters + PgBouncer"]
            direction TB
            PGB_ID["Deployment pgbouncer-identity"]:::infra
            PG_ID[("StatefulSet postgres-identity<br/>pg_identity + PVC")]:::infra
            PGB_PAY["Deployment pgbouncer-payment"]:::infra
            PG_PAY[("StatefulSet postgres-payment<br/>pg_payment + PVC")]:::infra
            PGB_FIN["Deployment pgbouncer-financial"]:::infra
            PG_FIN[("StatefulSet postgres-financial<br/>pg_financial + PVC")]:::infra
        end

        REDIS_CLUSTER[("Redis Cluster + PVC")]:::infra
        KAFKA[("Kafka StatefulSet")]:::infra
        CLICKHOUSE[("ClickHouse + PVC<br/>schema ConfigMap")]:::infra

        subgraph ObsPods["Observability"]
            PROM["Prometheus"]:::obs
            GRAFANA["Grafana"]:::obs
            LOKI["Loki + PVC"]:::obs
            PROMTAIL["Promtail DaemonSet"]:::obs
            JAEGER["Jaeger"]:::obs
            OTEL["OTel Collector"]:::obs
            NODEX["Node Exporter DaemonSet"]:::obs
            ALERTMGR["Alertmanager"]:::obs
            PYRO["Pyroscope"]:::obs
        end

        subgraph Jobs["Workers"]
            EMAILJ["email worker pod"]:::job
            AISP["ai-security pod"]:::job
        end
    end

    REPO --> ARGO
    ARGO --> K8S

    NGINX --> APIGW
    APIGW -->|"gRPC"| CorePods
    APIGW -->|"gRPC"| SREADER
    CorePods --> DataPods
    CorePods --> REDIS_CLUSTER
    CorePods --> KAFKA
    KAFKA --> EMAILJ
    KAFKA --> AISP
    KAFKA --> SWRITER
    SWRITER --> CLICKHOUSE
    SREADER --> CLICKHOUSE

    PGB_ID --> PG_ID
    PGB_PAY --> PG_PAY
    PGB_FIN --> PG_FIN

    CorePods -.->|"/metrics"| PROM
    CorePods -.->|"OTLP"| OTEL
    OTEL -.-> JAEGER
    PROMTAIL -.-> LOKI
    PROM -.-> GRAFANA
    PROM -.-> ALERTMGR
```

---

## Technology Stack

| Category | Technology | Purpose |
| :--- | :--- | :--- |
| **Language** | Go (Golang) + Python | Go for services, Python for `ai-security` |
| **API Edge Gateway** | Echo (REST) | REST API gateway with auto-generated Swagger UI |
| **RPC Inter-service** | gRPC + Protobuf | Contract-first synchronous communication |
| **OLTP Database** | PostgreSQL 17 ×3 | Database-per-bounded-context: `pg_identity`, `pg_payment`, `pg_financial` |
| **Connection Pooler** | PgBouncer ×3 | One pooler per cluster — `transaction` locally, `session` in K8s — host ports `6432`–`6434` |
| **OLAP Database** | ClickHouse | Columnar warehouse for analytics aggregations |
| **ORM** | GORM | Object-relational mapping & query builder |
| **DB Migrations** | Goose | Prefix-aware migrations applied at service startup |
| **Caching Tier** | Redis Cluster | 6-node cluster with per-context prefixes (`REDIS_1`–`REDIS_3`) |
| **Messaging Stream** | Apache Kafka (KRaft) | Asynchronous high-throughput event bus |
| **Token Manager** | JWT | Stateless authentication & authorization |
| **Observability** | OpenTelemetry + Jaeger | Vendor-neutral telemetry pipeline and visualization |
| **Metrics / Dashboards** | Prometheus + Grafana | Scraping, dashboards, and alert rules |
| **Log Aggregation** | Loki + Promtail | Centralized structured log storage and shipping |
| **Continuous Profiler** | Pyroscope | Real-time CPU/memory profiling |
| **Alerting** | Alertmanager | Alert routing & notification dispatch |
| **Reverse Proxy** | NGINX | Edge routing and TLS termination |
| **Containerization** | Docker + Docker Compose | Local orchestration |
| **Orchestrator** | Kubernetes + HPA | Production auto-scaling pod infrastructure |
| **GitOps** | ArgoCD | Declarative continuous delivery |
| **Load Testing** | k6 | Performance and load test suites |
| **E2E Testing** | Hurl | HTTP endpoint E2E suites |
| **Resilience** | `pkg/resilience` | Circuit breaker, rate limiter, load monitor, `DependencyGuard` |

---

## Getting Started

### Prerequisites

- [Git](https://git-scm.com/)
- [Go](https://go.dev/) (v1.23+)
- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
- [Just Task Runner](https://github.com/casey/just) or standard `make`
- [Protobuf Compiler](https://grpc.io/docs/protoc-installation/) (for codegen updates)

### 1. Clone the Workspace

```sh
git clone https://github.com/MamangRust/microservice-payment-gateway-grpc.git
cd microservice-payment-gateway-grpc
```

### 2. Prepare Environment Configurations

```sh
cp .env.example .env
cp deployments/local/docker.env.example deployments/local/docker.env
```

The PostgreSQL cluster contract lives there:

```dotenv
# Host = the context's PgBouncer service, never PostgreSQL itself
DB_IDENTITY_HOST=pgbouncer_identity
DB_IDENTITY_PORT=5432
DB_IDENTITY_NAME=pg_identity

DB_PAYMENT_HOST=pgbouncer_payment
DB_PAYMENT_PORT=5432
DB_PAYMENT_NAME=pg_payment

DB_FINANCIAL_HOST=pgbouncer_financial
DB_FINANCIAL_PORT=5432
DB_FINANCIAL_NAME=pg_financial
```

For **native** runs against `docker-compose.infra.yml`, point the hosts at `localhost` and the
ports at the published pooler ports `6432`–`6434`.

### 3. Start Local Environment

```sh
make build-up      # or: just build-up
```

Database migrations are **not** a separate step: each service applies its own Goose migrations at
startup against its bounded-context database, selected by the cluster prefix in
`service/<name>/cmd/main.go`.

```sh
make ps            # or: just ps
```

### 4. Access Services

| Service | URL |
| :--- | :--- |
| Swagger UI | [http://localhost/swagger/index.html](http://localhost/swagger/index.html) |
| REST API Gateway Edge | [http://localhost/api/*](http://localhost/api/*) |
| API Gateway Direct | [http://localhost:5000](http://localhost:5000) |
| Grafana | [http://localhost:3000](http://localhost:3000) (`admin`/`admin`) |
| Prometheus | [http://localhost:9090](http://localhost:9090) |
| Jaeger | [http://localhost:16686](http://localhost:16686) |
| Pyroscope | [http://localhost:4040](http://localhost:4040) |

```sh
make down          # or: just down
```

---

## Port Map Registry

| Application / Service | Port / URL |
| :--- | :--- |
| **auth** gRPC | `50051` |
| **role** gRPC | `50052` |
| **card** gRPC | `50053` |
| **merchant** gRPC | `50054` |
| **user** gRPC | `50055` |
| **saldo** gRPC | `50056` |
| **topup** gRPC | `50057` |
| **transaction** gRPC | `50058` |
| **transfer** gRPC | `50059` |
| **withdraw** gRPC | `50060` |
| **stats-reader** gRPC | `50062` |
| **ai-security** gRPC | `50051` |
| **PgBouncer — identity** (`pg_identity`) | `localhost:6432` |
| **PgBouncer — payment** (`pg_payment`) | `localhost:6433` |
| **PgBouncer — financial** (`pg_financial`) | `localhost:6434` |
| **Redis Cluster nodes** | `redis-node-1` … `redis-node-6` (`:6379`) |
| **Kafka** | `kafka-1:9092` |
| **ClickHouse native / HTTP** | `localhost:9000` / `localhost:8123` |

> PostgreSQL is **not** published to the host — connect through the PgBouncer ports above.

---

## Makefile / Justfile Reference

| Target | Scope |
| :--- | :--- |
| `build-up` / `just build-up` | Rebuild service images and launch Compose |
| `up` / `just up` | Launch Compose without rebuilding |
| `down` / `just down` | Stop Compose stacks |
| `ps` / `just ps` | Health, uptime, and mapped ports |
| `generate-proto` / `just generate-proto` | Compile `.proto` into Go models under `pb/` |
| `generate-swagger` / `just generate-swagger` | Regenerate OpenAPI/Swagger specs |
| `test-auth` / `just test-auth` | Integration tests for the `auth` module |
| `just test-unit` | Unit tests under `pkg/` |
| `just test-integration` | Integration tests under `tests/` |
| `just test-all` | Unit + integration |
| `just build` | Compile all services into `bin/` |
| `just tidy-all` | `go mod tidy` across all modules |

---

## Workspace Directory Tree

```
microservice-payment-gateway-grpc/
├── proto/                          # Protobuf contracts
│   ├── role/                       #   Role query specifications
│   ├── user_role/                  #   UserRole create/delete/find-by-user
│   ├── card/                       #   Virtual card specifications
│   ├── saldo/                      #   Balance specifications
│   ├── topup/                      #   Funding specifications
│   ├── transaction/                #   Audit register specifications
│   ├── transfer/                   #   P2P transfer specifications
│   ├── withdraw/                   #   Settlement specifications
│   ├── merchant/                   #   Merchant declarations
│   ├── merchant_document/          #   Verification documents
│   ├── stats/                      #   OLAP analytics query contracts
│   ├── ai_security/                #   Fraud detection contracts
│   ├── user/ auth/                 #   Identity contracts
│   └── common/                     #   Shared protobuf types
├── pb/                             # Compiled protobuf outputs
├── shared/                         # Consolidated workspace module
│   ├── domain/                     #   Domain models & requests
│   ├── mapper/                     #   Proto ↔ Go converters
│   ├── cache/                      #   Redis caching wrappers
│   ├── observability/              #   Cache/tracing interceptors
│   ├── errors/                     #   Error templates
│   └── errorhandler/               #   Error handling utilities
├── pkg/                            # Platform libraries module
│   ├── adapter/                    #   Guarded gRPC adapters (role, user_role)
│   ├── auth/                       #   JWT utilities
│   ├── database/                   #   Prefix-aware GORM + Goose + error mapping + names.go
│   ├── clickhouse/                 #   OLAP connection factory
│   ├── kafka/                      #   Kafka producer/consumer
│   ├── redis/                      #   Cluster-aware Redis connectors
│   ├── resilience/                 #   Circuit breaker, rate limiter, load monitor
│   ├── server/                     #   gRPC bootstrap
│   ├── otel/                       #   Telemetry hooks
│   ├── logger/                     #   Zap structured logging
│   ├── randomvcc/                  #   VCC issuance helpers
│   ├── rupiah/                     #   IDR currency helpers
│   └── api-key/                    #   API key generation & validation
├── service/                        # Business domains
│   ├── apigateway/                 #   REST API gateway (Echo)
│   ├── auth/ user/ role/           #   Identity context
│   ├── card/ merchant/ saldo/      #   Payment context
│   ├── topup/ transaction/         #   Financial context
│   │   transfer/ withdraw/
│   ├── stats-writer/               #   Kafka → ClickHouse (OLAP write)
│   ├── stats-reader/               #   ClickHouse → gRPC (OLAP read)
│   ├── email/                      #   Kafka notification worker
│   └── ai-security/                #   Python fraud detection worker
├── deployments/
│   ├── local/                      #   docker-compose.yml + docker-compose.infra.yml
│   ├── kubernetes/                 #   Database, cache, messaging, networking,
│   │                               #   observability, security, services, overlays
│   └── gitops/argocd/              #   ArgoCD project, root app, production app
├── seeder/                         #   Database seeder
├── hurl/                           #   Hurl E2E HTTP suites
├── k6/                             #   k6 load test suites
├── observability/                  #   Prometheus, Loki, OTel, Promtail configs
├── grafana/                        #   Pre-configured dashboard templates
├── nginx/                          #   Reverse-proxy edge rules
├── redis/                          #   Redis cluster configuration
├── tests/                          #   Integration tests
└── images/                         #   Architecture diagrams & dashboards
```

---

## License

This project is open-sourced under the MIT License for educational and development purposes.

---

<p align="center">
  Built with Go, gRPC, Apache Kafka, ClickHouse OLAP, PostgreSQL clusters behind PgBouncer, and a passion for high-performance financial microservices.
</p>
