# Distributed Microservices Architecture — Payment Gateway Platform

A production-grade, highly resilient, and fully observable **microservices payment gateway backend** built in **Go (Golang)**. Designed around domain-driven service boundaries following Clean Architecture principles, it distributes financial workloads across self-contained, micro-deployable services while maintaining high consistency and performance.

Each financial and identity business domain — Users, Roles, Cards, Merchants, Saldo, Topups, Transactions, Transfers, Withdrawals — lives in its own self-contained microservice. These services communicate synchronously via lightweight **gRPC** protocols and asynchronously using **Apache Kafka** event propagation, exposing a unified entry point through a **REST API Gateway** built with **Echo** and NGINX.

For high-performance analytical queries and business intelligence, the platform features a **CQRS-style OLAP & Analytics Layer** powered by **ClickHouse**. Real-time events from Kafka are consumed by **Stats Writer**, which processes and writes them to ClickHouse, while the **Stats Reader** gRPC service provides ultra-fast aggregations (monthly/yearly volumes, payment methods, transaction statuses) back to the REST API Gateway.

The platform is fortified with a **comprehensive observability suite** (Prometheus, Grafana, Loki, Jaeger, OpenTelemetry, Pyroscope), robust connection pooling via **PgBouncer**, **isolated Redis caching** with custom telemetry for each service, and Kubernetes configurations ready for production auto-scaling.

---

## Key Features

| Domain | Capabilities |
| :--- | :--- |
| **Auth & Users** | Secure registration, multi-factor login, stateless JWT access/refresh token lifecycle, password reset workflows, OTP email verification, and `GetMe` profile resolver. |
| **Roles & RBAC** | Custom permission configuration, granular access control matrices, and sub-second permission evaluation cached via Redis. |
| **Cards & VCC** | Virtual and debit card CRUD operations with soft-delete capabilities, card activation/suspension toggles, and multi-dimensional transaction analytics (daily/monthly/yearly topup, withdraw, transfer). |
| **Merchants** | Fully featured merchant onboarding, profile details management, business data registration, and merchant performance/transaction reports with full data restoration capabilities (soft delete & restore). |
| **Saldo (Balance)** | High-throughput, thread-safe real-time balance calculations, optimistic concurrency locks, and localized balances. |
| **Topup** | Balance loading ledger engine supporting multiple payment methods, detailed transactions logging, and soft-delete audit records. |
| **Transaction** | Centralized financial audit ledger collecting transaction events across the system, global search filters, status tracking, and monthly/yearly volume reports. |
| **Transfer** | Safe peer-to-peer card-to-card or user-to-user funds settlement with balance debit/credit synchronization and event-driven logging. |
| **Withdraw** | Funds settlement from user cards to external accounts/banks, daily transaction threshold limits, and status processing pipelines. |
| **OLAP Analytics** | High-performance real-time data warehouse using **ClickHouse** featuring specialized **Stats Writer** (Kafka events consumer) and **Stats Reader** (gRPC analytics query server) services. |
| **Email Worker** | Kafka-driven asynchronous worker dispatching critical notification emails (OTPs, login alerts, merchant onboarding notices, and transfer/topup invoices) via SMTP. |
| **Observability** | Multi-dimensional metrics (Prometheus + Grafana), log aggregation (Loki + Promtail), end-to-end distributed tracing (Jaeger + OpenTelemetry), continuous CPU/Memory profiling (Pyroscope), and resource monitors (Node, Kafka, Postgres, ClickHouse Exporters). |
| **Deployment** | Local orchestration using Docker Compose (featuring individual Redis instances, ClickHouse server, and PgBouncer), and auto-scaling Kubernetes manifests configured with Horizontal Pod Autoscalers (HPA). |

---

## Architecture Overview

The platform implements a **Distributed Microservices Architecture**. Each business service is logical, decoupled, and self-contained inside the `service/` directory, possessing its own independent gRPC boundary. An **API Gateway** (Echo + NGINX) acts as the unified edge router, routing and translating client REST HTTP requests into fast gRPC downstream communications.

### Core Architecture Principles

- **Domain-Driven Boundary Isolation**: Every service owns its database access, caching layers, and service logic, strictly forbidding cross-boundary database sharing.
- **Clean Architecture**: Standardized layers of `handler → service → repository` ensure that business domains remain unaffected by framework or database changes.
- **PgBouncer Pooling**: Employs connection pooling to avoid PostgreSQL socket exhaustion across the multiple concurrent microservices.
- **OLAP & CQRS Pattern**: Transactional (OLTP) writes are processed in PostgreSQL, while analytical (OLAP) read queries are redirected to a dedicated ClickHouse database populated asynchronously via Kafka events.
- **Event-Driven Resilience**: Apache Kafka decouples transaction events, ensuring side effects like email billing and stats writing remain completely non-blocking.
- **OTel Telemetry Integration**: Standardized OpenTelemetry middleware injects trace IDs across gRPC boundaries, allowing seamless trace propagation from the client REST HTTP edge down to postgres and ClickHouse operations.

```mermaid
graph TB
    classDef client fill:#0f172a,stroke:#38bdf8,color:#e0f2fe,stroke-width:2px,font-weight:bold
    classDef gateway fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px,font-weight:bold
    classDef domain fill:#1e1b4b,stroke:#818cf8,color:#e0e7ff,stroke-width:1.5px
    classDef infra fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef obs fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1.5px
    classDef event fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px
    classDef olap fill:#1e293b,stroke:#a855f7,color:#f3e8ff,stroke-width:1.5px

    Client["Client Applications<br/>(Web / Mobile / API)"]:::client

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
            AUTH["Auth Service<br/>JWT & OTP Verification"]
            USER["User Service<br/>Profile Management"]
            ROLE["Role Service<br/>RBAC & Permissions"]
        end

        subgraph MerchantDomain["Merchant Management"]
            MERCH["Merchant Service<br/>Onboarding & Profiling"]
        end

        subgraph FinanceDomain["Finance & Ledger Suite"]
            CARD["Card Service<br/>VCC & Card Analytics"]
            SALDO["Saldo Service<br/>Real-time Balance Tracker"]
        end

        subgraph TransactionDomain["Transfers & Transactions"]
            TOPUP["Topup Service<br/>Balance Funding Engine"]
            TXN["Transaction Service<br/>Central Audit Register"]
            TRANSFER["Transfer Service<br/>P2P Card-to-Card Transfer"]
            WITHDRAW["Withdraw Service<br/>Outbound Fund Settlement"]
        end
    end
    class BusinessServices domain

    subgraph OLAPEngine["OLAP & Analytics Layer"]
        direction TB
        WRITER["Stats Writer Service<br/>Kafka Event Consumer"]:::olap
        READER["Stats Reader Service<br/>gRPC Query Service"]:::olap
        CLICKHOUSE[("ClickHouse OLAP<br/>Analytics DB")]:::infra
    end

    APIGateway -->|"gRPC (Core OLTP)"| BusinessServices
    APIGateway -->|"gRPC (OLAP Queries)"| READER

    subgraph Infrastructure["Infrastructure Layer"]
        direction LR
        PGBOUNCER["PgBouncer<br/>Connection Pooler :6432"]
        PG[("PostgreSQL<br/>PAYMENT_GATEWAY DB")]
        REDIS[("Redis Cache Cluster<br/>11 Isolated Databases")]
        KAFKA[("Kafka<br/>Event Bus")]
        PYRO["Pyroscope<br/>Continuous Profiler"]
    end
    class Infrastructure infra

    BusinessServices -->|"gRPC / SQL"| PGBOUNCER
    PGBOUNCER --> PG
    BusinessServices -->|"Cache / Invalidate"| REDIS
    BusinessServices -->|"Publish Events"| KAFKA
    BusinessServices -.->|"Profile Data"| PYRO

    subgraph EventConsumers["Event-Driven Consumers"]
        EMAIL["Email Service<br/>SMTP Notification Worker"]
    end
    class EventConsumers event

    KAFKA -->|"Consume Events"| EMAIL
    KAFKA -->|"Consume Events"| WRITER
    WRITER -->|"Insert Events"| CLICKHOUSE
    READER -->|"Query Stats"| CLICKHOUSE
    READER -->|"Cache Stats"| REDIS

    subgraph Observability["Observability Stack"]
        direction LR
        PROM["Prometheus<br/>Metrics Engine"]
        LOKI["Loki<br/>Log Aggregator"]
        JAEGER["Jaeger<br/>Distributed Traces"]
        GRAFANA["Grafana<br/>Unified Dashboards"]
        OTEL["OTel Collector<br/>Telemetry Pipeline"]
        PROMTAIL["Promtail<br/>Log Shipper"]
        NODEX["Node Exporter<br/>System Metrics"]
        KAFKAX["Kafka Exporter<br/>Topic lag / Broker health"]
        PGX["Postgres Exporter<br/>DB Performance"]
    end
    class Observability obs

    BusinessServices -.->|"/metrics"| PROM
    BusinessServices -.->|"Traces"| OTEL
    READER -.->|"/metrics"| PROM
    WRITER -.->|"/metrics"| PROM
    OTEL -.-> JAEGER
    PROMTAIL -.-> LOKI
    NODEX -.-> PROM
    KAFKAX -.-> PROM
    PGX -.-> PROM
    PROM -.-> GRAFANA
    LOKI -.-> GRAFANA
    JAEGER -.-> GRAFANA
```

---

## Service Catalog

The microservices architecture consists of **14 logical micro-services/workers** plus supporting database and migrations:

```mermaid
graph LR
    classDef svc fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1px,rx:8
    classDef gw fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px,rx:8,font-weight:bold
    classDef support fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1px,rx:8
    classDef olap fill:#1e293b,stroke:#a855f7,color:#f3e8ff,stroke-width:1px,rx:8

    subgraph Gateway
        API["API Gateway<br/>Echo + REST + Swagger"]:::gw
    end

    subgraph Identity["Identity & Access (3)"]
        A1["auth"]:::svc
        A2["user"]:::svc
        A3["role"]:::svc
    end

    subgraph Merchant["Merchant Suite (1)"]
        M1["merchant"]:::svc
    end

    subgraph Finance["Finance & Card Suite (2)"]
        F1["card"]:::svc
        F2["saldo"]:::svc
    end

    subgraph Movements["Fund Transactions (4)"]
        T1["topup"]:::svc
        T2["transaction"]:::svc
        T3["transfer"]:::svc
        T4["withdraw"]:::svc
    end

    subgraph OLAP["OLAP & Analytics (2)"]
        O1["stats-writer"]:::olap
        O2["stats-reader"]:::olap
    end

    subgraph Support["Support Services (2)"]
        S1["email"]:::support
        S2["migrate"]:::support
    end

    API --> Identity
    API --> Merchant
    API --> Finance
    API --> Movements
    API --> OLAP
```

---

## Internal Service Architecture

Every logical business service is mapped as a decoupled submodule following structured clean architecture rules.

```mermaid
graph TB
    classDef handler fill:#1e3a5f,stroke:#7dd3fc,color:#e0f2fe,stroke-width:1.5px
    classDef service fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1.5px
    classDef repo fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef infra fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1.5px
    classDef shared fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px

    subgraph Service["service/<name>/"]
        direction TB

        CMD["cmd/main.go or main.go<br/>Entry Point"]

        subgraph Internal["internal/"]
            direction TB
            APPS["app/client.go or server.go<br/>Dependency Wiring"]:::handler
            HANDLER["handler/<br/>gRPC Handlers"]:::handler
            MW["middleware/<br/>Interceptors"]:::handler
            SVC["service/<br/>Business Logic"]:::service
            CACHE["cache/<br/>Redis Cache Layer"]:::service
            REPO["repository/<br/>Data Access (GORM)"]:::repo
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
        DOMAIN["domain/<br/>record / requests / response"]:::shared
        OBS["observability/<br/>cache_metrics / tracing_metrics"]:::shared
        CACHESHARED["cache/<br/>redis_cache.go"]:::shared
        PB["pb/<br/>Protobuf Generated Code"]:::shared
        MAPPER["mapper/<br/>Domain ↔ Proto"]:::shared
        ERRORS["errors/ + errorhandler/"]:::shared
    end

    subgraph PkgLibs["pkg/ — Platform Libraries"]
        direction LR
        PKGAUTH["auth/<br/>JWT Manager"]:::infra
        PKGKAFKA["kafka/<br/>Producer / Consumer"]:::infra
        PKGOTEL["otel/<br/>Tracing + Metrics Init"]:::infra
        PKGRES["resilience/<br/>Circuit Breaker<br/>Rate Limiter<br/>Load Monitor"]:::infra
        PKGLOG["logger/<br/>Zap Structured Logging"]:::infra
        PKGSRV["server/<br/>gRPC Server Bootstrap"]:::infra
        PKGDB["database/<br/>PostgreSQL connection"]:::infra
    end

    REPO --> DOMAIN
    REPO --> PB
    SVC --> DOMAIN
    SVC --> OBS
    HANDLER --> PB
    HANDLER --> MAPPER
    APPS --> PKGSRV
    APPS --> PKGOTEL
    APPS --> CACHESHARED
    APPS --> OBS
```

---

## Data & Event Flow

### Synchronous Flow (gRPC & Cache Read-Through)

All external client API requests go through REST API HTTP requests. The API Gateway authorizes the JWT, connects with the correct downstream gRPC modular server, checks Redis caching, and fetches PostgreSQL through PgBouncer if a cache miss happens.

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant GW as API Gateway<br/>(Echo + REST)
    participant SVC as Domain Service<br/>(gRPC Server)
    participant REDIS as Redis Database
    participant PGB as PgBouncer
    participant DB as PostgreSQL

    C->>GW: REST HTTP Request (GET/POST/PUT/DELETE)
    GW->>GW: JWT Authentication Check
    GW->>SVC: gRPC Call (Protobuf payload)
    SVC->>REDIS: Check Cache
    alt Cache Hit
        REDIS-->>SVC: Return Cached Response
    else Cache Miss
        SVC->>PGB: Acquire Connection
        PGB->>DB: SQL Query execution
        DB-->>PGB: DB Result Set
        PGB-->>SVC: SQL rows mapped
        SVC->>REDIS: Populate Cache for next read
    end
    SVC-->>GW: gRPC Response payload
    GW-->>C: REST HTTP Response (JSON format)
```

### Asynchronous Flow (Kafka Notification Event pipeline)

High-performance transaction modifications (like transfers or top-ups) trigger background notification events published directly to Apache Kafka brokers. The isolated Email service listens to Kafka, maps the events, and contacts Ethereal/SMTP services.

```mermaid
sequenceDiagram
    autonumber
    participant SVC as Transaction / Topup / Transfer
    participant K as Kafka Broker
    participant EMAIL as Email Worker Service
    participant SMTP as SMTP Server

    SVC->>K: Publish Event (e.g. transfer.created / topup.success)
    K-->>EMAIL: Deliver topic payload (asynchronous consumer)
    EMAIL->>EMAIL: Map payload details
    EMAIL->>SMTP: Send custom styled notification
    SMTP-->>EMAIL: Delivery Confirmation
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
        SVCS["All Business Services<br/>(11 services)"]:::service
        KAFKA_SRC["Kafka Broker"]:::service
        NODES["Host / Node"]:::service
        DB_SRC["PostgreSQL Engine"]:::service
    end

    subgraph Collectors["Collection Layer"]
        direction TB
        PROM["Prometheus<br/>Scrapes /metrics"]:::collector
        PROMTAIL["Promtail<br/>Ships container logs"]:::collector
        OTEL["OTel Collector<br/>Receives OTLP spans"]:::collector
        NODEX["Node Exporter<br/>CPU / Memory / Disk / Net"]:::collector
        KAFKAX["Kafka Exporter<br/>Topic lag / Broker health"]:::collector
        PGX["Postgres Exporter<br/>PgBouncer & Query performance"]:::collector
    end

    subgraph Storage["Storage Layer"]
        direction TB
        PROM_TSDB["Prometheus TSDB<br/>(Metrics)"]:::storage
        LOKI_STORE["Loki<br/>(Log Index + Chunks)"]:::storage
        JAEGER_STORE["Jaeger<br/>(Trace Storage)"]:::storage
    end

    subgraph Visualization["Visualization & Alerting"]
        GRAFANA["Grafana<br/>Unified Dashboards"]:::viz
        ALERTMGR["Alertmanager<br/>Alert Routing"]:::viz
    end

    SVCS -->|"/metrics"| PROM
    SVCS -->|"OTLP gRPC"| OTEL
    SVCS -->|"stdout/stderr"| PROMTAIL
    NODES --> NODEX
    KAFKA_SRC --> KAFKAX
    DB_SRC --> PGX

    NODEX --> PROM
    KAFKAX --> PROM
    PGX --> PROM
    PROM --> PROM_TSDB
    PROMTAIL --> LOKI_STORE
    OTEL --> JAEGER_STORE

    PROM_TSDB --> GRAFANA
    LOKI_STORE --> GRAFANA
    JAEGER_STORE --> GRAFANA
    PROM_TSDB --> ALERTMGR
```

| Pillar | Tool | Purpose |
| :--- | :--- | :--- |
| **Metrics** | Prometheus + Grafana | Core metrics tracking (CPU, memory, request error rates, gRPC latencies, DB connection states). |
| **Logging** | Loki + Promtail | Centralized structured JSON logger for indexing logs by service, queryable via LogQL. |
| **Tracing** | OpenTelemetry + Jaeger | Distributed system tracing across API gateway and internal gRPC services. |
| **Profiling** | Pyroscope | Continuous memory/CPU profiling to eliminate allocation memory leaks in transaction loops. |
| **Alerting** | Alertmanager | Automated notification system triggered during latency hikes or service disconnects. |

---

## Deployment Architectures

### Docker Compose (Local Development)

The Docker Compose configuration provisions 11 isolated database setups inside a single containerized environment to replicate real microservices patterns.

```mermaid
flowchart TD
    classDef gateway fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px,font-weight:bold
    classDef core fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1.5px
    classDef infra fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef obs fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1.5px
    classDef event fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px
    classDef olap fill:#1e293b,stroke:#a855f7,color:#f3e8ff,stroke-width:1.5px

    subgraph DockerCompose["docker-compose.yml — Local Environment"]

        subgraph Gateway["API Gateway"]
            NGINX["NGINX Proxy :80"]
            APIGW["API Gateway Container<br/>Echo + REST :5000"]
        end
        class Gateway gateway

        subgraph Services["Core Service Containers"]
            subgraph Identity["Identity & Access"]
                AUTH["auth-service"]
                USER["user-service"]
                ROLE["role-service"]
            end

            subgraph MerchantSuite["Merchant Domain"]
                MERCH["merchant-service"]
            end

            subgraph FinanceSuite["Finance & Card"]
                CARD["card-service"]
                SALDO["saldo-service"]
            end

            subgraph MovementsSuite["Fund Movements"]
                TOPUP["topup-service"]
                TXN["transaction-service"]
                TRANSFER["transfer-service"]
                WITHDRAW["withdraw-service"]
            end

            subgraph OLAPSuite["OLAP & Analytics"]
                STATS_WRITER["stats-writer"]:::olap
                STATS_READER["stats-reader"]:::olap
            end
        end
        class Services core

        subgraph Infra["Infrastructure Suite"]
            PG[("PostgreSQL :5432")]
            PGB[("PgBouncer :6432")]
            REDIS_APIGW[("redis-apigateway :6379")]
            REDIS_AUTH[("redis-auth :6380")]
            REDIS_USER[("redis-user :6381")]
            REDIS_CARD[("redis-card :6382")]
            REDIS_MERCH[("redis-merchant :6383")]
            REDIS_ROLE[("redis-role :6384")]
            REDIS_SALDO[("redis-saldo :6385")]
            REDIS_TXN[("redis-transaction :6386")]
            REDIS_TOPUP[("redis-topup :6387")]
            REDIS_TRANS[("redis-transfer :6388")]
            REDIS_WITHDRAW[("redis-withdraw :6389")]
            KAFKA[("Kafka Broker :9092")]
            CLICKHOUSE[("ClickHouse OLAP :9000/:8123")]
            PYRO[("Pyroscope :4040")]
        end
        class Infra infra

        subgraph Obs["Observability Stack"]
            PROM["Prometheus :9090"]
            GRAFANA["Grafana :3000"]
            LOKI["Loki :3100"]
            JAEGER["Jaeger :16686"]
            OTEL["OTel Collector :4317"]
            NODEX["Node Exporter"]
            KAFKAX["Kafka Exporter"]
            PGX["Postgres Exporter"]
            PROMTAIL["Promtail Log Shipper"]
        end
        class Obs obs

        subgraph Events["Event Consumers"]
            EMAIL["Email Worker"]
        end
        class Events event
    end

    NGINX --> APIGW
    APIGW -->|"gRPC"| Services
    APIGW -->|"gRPC"| STATS_READER
    Services -->|"gRPC/SQL"| PGB
    PGB --> PG
    Services --> KAFKA
    KAFKA --> EMAIL
    KAFKA --> STATS_WRITER
    STATS_WRITER --> CLICKHOUSE
    STATS_READER --> CLICKHOUSE

    AUTH --> REDIS_AUTH
    USER --> REDIS_USER
    CARD --> REDIS_CARD
    MERCH --> REDIS_MERCH
    ROLE --> REDIS_ROLE
    SALDO --> REDIS_SALDO
    TOPUP --> REDIS_TOPUP
    TXN --> REDIS_TXN
    TRANSFER --> REDIS_TRANS
    WITHDRAW --> REDIS_WITHDRAW
    APIGW --> REDIS_APIGW

    Services -.->|"Metrics"| PROM
    Services -.->|"Traces"| OTEL
    Services -.->|"Profiles"| PYRO
    STATS_WRITER -.->|"Metrics"| PROM
    STATS_READER -.->|"Metrics"| PROM
    STATS_WRITER -.->|"Traces"| OTEL
    STATS_READER -.->|"Traces"| OTEL
    OTEL -.-> JAEGER
    PROMTAIL -.-> LOKI
    PROM -.-> GRAFANA
    LOKI -.-> GRAFANA
```

### Kubernetes (Production Ready)

Our enterprise Kubernetes infrastructure resides inside the dedicated `payment-gateway` namespace, configuring scalable nodes utilizing Horizontal Pod Autoscalers.

```mermaid
flowchart TD
    classDef k8s fill:#0c1222,stroke:#38bdf8,color:#e0f2fe,stroke-width:2px,font-weight:bold
    classDef pod fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1.5px
    classDef hpa fill:#3b0764,stroke:#c084fc,color:#f3e8ff,stroke-width:1px,font-style:italic
    classDef infra fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef obs fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1.5px
    classDef job fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px
    classDef olap fill:#1e293b,stroke:#a855f7,color:#f3e8ff,stroke-width:1.5px

    subgraph K8S["Kubernetes Cluster — namespace: payment-gateway"]

        subgraph Ingress["Ingress Controller"]
            NGINX["NGINX Ingress<br/>+ TLS Termination"]:::k8s
            APIGW["API Gateway Pod"]:::pod
        end

        subgraph CorePods["Core Service Pods + HPAs"]
            direction TB

            subgraph IdentityPods["Identity Suite"]
                AUTH["auth-pod"]:::pod
                USER["user-pod"]:::pod
                ROLE["role-pod"]:::pod
                AUTH_HPA["↕ HPA"]:::hpa
                USER_HPA["↕ HPA"]:::hpa
            end

            subgraph MerchPods["Merchant Domain"]
                MERCH["merchant-pod"]:::pod
            end

            subgraph FinancePods["Finance & Ledger"]
                CARD["card-pod"]:::pod
                SALDO["saldo-pod"]:::pod
            end

            subgraph TransPods["Transaction Engine"]
                TOPUP["topup-pod"]:::pod
                TXN["transaction-pod"]:::pod
                TRANSFER["transfer-pod"]:::pod
                WITHDRAW["withdraw-pod"]:::pod
            end

            subgraph OLAPPods["OLAP Analytics"]
                STATS_WRITER["stats-writer-pod"]:::olap
                STATS_READER["stats-reader-pod"]:::olap
            end
        end

        subgraph InfraPods["Infrastructure Pods"]
            PG[("PostgreSQL DB + PVC")]:::infra
            PGB["PgBouncer Daemon"]:::infra
            REDIS_CLUSTER[("Redis Cluster + PVC")]:::infra
            KAFKA[("Kafka StatefulSet")]:::infra
            CLICKHOUSE[("ClickHouse OLAP + PVC")]:::infra
        end

        subgraph ObsPods["Observability Pods"]
            PROM["Prometheus Pod"]:::obs
            GRAFANA["Grafana Pod"]:::obs
            LOKI["Loki Pod + PVC"]:::obs
            PROMTAIL["Promtail DaemonSet"]:::obs
            JAEGER["Jaeger Pod"]:::obs
            OTEL["OTel Collector Pod"]:::obs
            NODEX["Node Exporter DaemonSet"]:::obs
            ALERTMGR["Alertmanager Pod"]:::obs
            PYRO["Pyroscope Pod"]:::obs
        end

        subgraph Jobs["Jobs & Workers"]
            MIGRATE["Migration Job"]:::job
            EMAIL["Email Worker Pod"]:::job
        end
    end

    NGINX --> APIGW
    APIGW -->|"gRPC"| CorePods
    APIGW -->|"gRPC"| STATS_READER
    CorePods --> PGB
    PGB --> PG
    CorePods --> REDIS_CLUSTER
    CorePods --> KAFKA
    KAFKA --> EMAIL
    KAFKA --> STATS_WRITER
    STATS_WRITER --> CLICKHOUSE
    STATS_READER --> CLICKHOUSE

    CorePods -.->|"/metrics"| PROM
    CorePods -.->|"OTLP"| OTEL
    STATS_WRITER -.->|"/metrics"| PROM
    STATS_READER -.->|"/metrics"| PROM
    STATS_WRITER -.->|"OTLP"| OTEL
    STATS_READER -.->|"OTLP"| OTEL
    OTEL -.-> JAEGER
    PROMTAIL -.-> LOKI
    PROM -.-> GRAFANA
    MIGRATE --> PG
```

---

## Technology Stack

| Category | Selected Technologies | Purpose |
| :--- | :--- | :--- |
| **Language** | Go (Golang v1.25) | High-performance compiled concurrent backend execution. |
| **API Edge Gateway** | Echo (REST API) | High-performance REST API Gateway engine with auto-generated Swagger UI. |
| **RPC Inter-service** | gRPC + Protobuf | Blazing fast, contract-first synchronous communications. |
| **OLTP Database** | PostgreSQL v17 | Safe ACID ledger persistent storage system. |
| **OLAP Database** | ClickHouse | Ultra-high performance column-oriented data warehouse for aggregations and analytics. |
| **Database Gateway**| PgBouncer | Extreme-efficiency PostgreSQL socket connection pooler. |
| **ORM** | GORM | Object-relational mapping & query builder on PostgreSQL. |
| **DB Migrations** | Goose | Incremental database schema version manager. |
| **Caching Tier** | Redis | Multi-database low-latency key-value cached engine. |
| **Messaging Stream** | Apache Kafka | Asynchronous high-throughput messaging event bus. |
| **Token Manager** | JWT | Secure stateless request authentication standard. |
| **Observability** | OpenTelemetry + Jaeger | Vendor-neutral distributed telemetry pipeline and visualization. |
| **Continuous Profiler**| Pyroscope | Real-time memory allocation tracker to identify hot paths. |
| **Docker Engine** | Compose | Local environment virtualization orchestration. |
| **Orchestrator** | Kubernetes | Production-scale auto-scaling pod clustering infrastructure. |

---

## Getting Started

### Prerequisites

Ensure the following system packages are locally configured:

- [Git](https://git-scm.com/)
- [Go](https://go.dev/) (v1.23+)
- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
- [Just Task Runner](https://github.com/casey/just) or Standard `Make`
- [Protobuf Compiler](https://grpc.io/docs/protoc-installation/) (for codegen updates)

### 1. Clone the Workspace

```sh
git clone https://github.com/MamangRust/microservice-payment-gateway-grpc.git
cd microservice-payment-gateway-grpc
```

### 2. Prepare Environment Configurations

Setup the system configurations from placeholders:

```sh
# Copy root variables
cp .env.example .env

# Copy local docker settings overrides
cp deployments/local/docker.env.example deployments/local/docker.env
```

### 3. Start Local Environment (Docker Compose)

Launch all infrastructure utilities, telemetry containers, and application submodules:

```sh
# Compile Docker images & boot Compose layers
make build-up
# Or using Just:
just build-up

# Execute Goose migration scripts
make migrate
# Or using Just:
just migrate

# (Optional) Insert development mock data into DB
make seeder
# Or using Just:
just seeder
```

Ensure everything has booted successfully:

```sh
make ps
# Or using Just:
just ps
```

### 4. Port Map Registry

| Application/Service | Port Configuration / URL |
| :--- | :--- |
| **Swagger API Documentation** | [http://localhost/swagger/index.html](http://localhost/swagger/index.html) |
| **REST API Gateway Edge** | [http://localhost/api/*](http://localhost/api/*) |
| **API Gateway Direct** | [http://localhost:5000](http://localhost:5000) |
| **API Gateway Direct Swagger** | [http://localhost:5000/swagger/index.html](http://localhost:5000/swagger/index.html) |
| **Grafana Dashboard Portal** | [http://localhost:3000](http://localhost:3000) *(Credentials: `admin`/`admin`)* |
| **Prometheus Telemetry** | [http://localhost:9090](http://localhost:9090) |
| **Jaeger Distributed Tracing** | [http://localhost:16686](http://localhost:16686) |
| **Pyroscope Profiling Panel** | [http://localhost:4040](http://localhost:4040) |
| **PgBouncer Gateway Node** | `localhost:6432` |
| **PostgreSQL Database Engine** | `localhost:5432` |
| **Stats Reader gRPC Service** | `localhost:50062` |
| **ClickHouse Native TCP** | `localhost:9000` |
| **ClickHouse HTTP Interface** | `localhost:8123` |

To fully stop the development system:

```sh
make down
# Or using Just:
just down
```

---

## Makefile / Justfile Reference

The workspace includes a standard `Makefile` and `justfile` featuring mirroring tasks:

| Target/Recipe | Execution Scope |
| :--- | :--- |
| `build-up` / `just build-up` | Recompiles local service Dockerfiles and launches the compose stacks. |
| `up` / `just up` | Launches existing Compose containers without rebuilding images. |
| `down` / `just down` | Stops the local compose container stacks and releases networks. |
| `ps` / `just ps` | Displays health, uptime, and mapped ports of the container cluster. |
| `migrate` / `just migrate` | Triggers PostgreSQL schema updates via the migrate binary. |
| `migrate-down` / `just migrate-down` | Rolls back the latest database migration states. |
| `generate-proto` / `just generate-proto` | Compiles `.proto` files down into Go models within `/pb`. |
| `generate-swagger` / `just generate-swagger` | Regenerates OpenAPI/Swagger schema specifications. |
| `test-auth` / `just test-auth` | Executes integration tests for the `auth` module. |
| `just test-unit` | Executes Go standard library test routines under `pkg/`. |
| `just test-integration` | Triggers end-to-end integration flows under `tests/`. |
| `just test-all` | Chain-runs standard unit tests alongside integration suites. |
| `just build` | Locally compiles all sub-services into unified `bin/` folders. |
| `just tidy-all` | Iterates across all go modules, executing `go mod tidy` cleanups. |

---

## Workspace Directory Tree

```
microservice-payment-gateway-grpc/
├── proto/                          # Protobuf contracts (12 domains)
│   ├── auth.proto                  #   Identity tokens contracts
│   ├── card/                       #   Virtual Card specifications
│   ├── common/                     #   Shared protobuf data types
│   ├── merchant/                   #   Merchant account declarations
│   ├── merchant_document/          #   Verification files specifications
│   ├── role/                       #   Role mapping specifications
│   ├── saldo/                      #   Balance updates specifications
│   ├── topup/                      #   Funding balance specifications
│   ├── transaction/                #   General audit register specifications
│   ├── transfer/                   #   Peer-to-peer transaction specifications
│   ├── user/                       #   User CRUD data properties
│   └── withdraw/                   #   Bank settlement configurations
├── pb/                             # Compiled protobuf outputs
├── shared/                         # Consolidated workspace module
│   ├── domain/                     #   Internal domain models & requests
│   ├── mapper/                     #   Bidirectional converters (Proto ↔ Go)
│   ├── cache/                      #   Redis caching wrappers
│   ├── observability/              #   Cache/Tracing monitoring interceptors
│   ├── errors/                     #   Localized error templates
│   └── errorhandler/               #   Global transaction handlers
├── pkg/                            # Multi-project utilities module
│   ├── adapter/                    #   External connection controllers
│   ├── api-key/                    #   API Key generation & validation
│   ├── auth/                       #   JWT authentication utilities
│   ├── database/                   #   Relational DB initialization hooks
│   ├── date/                       #   Time utilities
│   ├── dotenv/                     #   Environment loader helper
│   ├── email/                      #   SMTP email driver
│   ├── hash/                       #   Bcrypt cryptography utilities
│   ├── kafka/                      #   Kafka writer/reader configurations
│   ├── logger/                     #   Structured Zap logs wrappers
│   ├── clickhouse/                 #   ClickHouse database connection factory
│   ├── method_topup/               #   Valid top-up methods definitions
│   ├── middleware/                 #   HTTP/gRPC general middleware
│   ├── otel/                       #   Telemetry hooks config
│   ├── random_string/              #   String helper generators
│   ├── randomvcc/                  #   VCC balance issuer helpers
│   ├── redis/                      #   Redis backend connectors
│   ├── resilience/                 #   Circuit breaker, rate limiter, load monitor
│   ├── rupiah/                     #   Indonesian Rupiah cash mappings
│   ├── server/                     #   gRPC bootstrap templates
│   └── trace_unic/                 #   Observability tracing utilities
├── service/                        # Isolated functional business domains
│   ├── apigateway/                 #   Unified REST API Gateway (Echo)
│   ├── auth/                       #   Identity authentication engine
│   ├── user/                       #   User profiles administration
│   ├── role/                       #   RBAC authorization configurations
│   ├── merchant/                   #   Merchants registration
│   ├── card/                       #   Debit & VCC virtual logs
│   ├── saldo/                      #   Real-time balance records
│   ├── topup/                      #   Funding balance processor
│   ├── transaction/                #   Central audit ledger service
│   ├── transfer/                   #   P2P funding transfer engine
│   ├── withdraw/                   #   Outbound settlement handler
│   ├── stats-writer/               #   ClickHouse Kafka events consumer (OLAP writer)
│   ├── stats-reader/               #   ClickHouse gRPC analytics queries server (OLAP reader)
│   ├── email/                      #   Asynchronous Kafka notification worker
│   └── migrate/                    #   Incremental DB migrations runner
├── deployments/
│   ├── local/                      #   Docker compose infrastructure scripts
│   └── kubernetes/                 #   Production K8s manifests (deployments, HPAs, volumes)
├── observability/                  #   Telemetry pipeline yaml profiles (Loki, OTEL, Promtail)
├── grafana/                        #   Pre-configured dashboard templates
├── nginx/                          #   Reverse-proxy edge rules
├── redis/                          #   Advanced Redis configs
└── images/                         #   Architecture diagrams & dashboard screenshots
```

---


## License

This project is open-sourced under the MIT License for educational and development purposes.

---

<p align="center">
  Built with Go, gRPC, Apache Kafka, ClickHouse OLAP, and a passion for high-performance microservices.
</p>