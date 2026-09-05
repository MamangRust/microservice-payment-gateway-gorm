## PKG

This repository contains reusable **utilities, modules, and helpers** designed to support a microservice-based architecture for a Digital Payment platform. These components are used across various services such as **Auth**, **User**, **Merchant**, **Topup**, **Withdraw**, and etc.


```
.
├── api-key # API Key generation and validation
│   ├── apikey.go
│   ├── apikey_test.go
│   └── README.md
├── auth # JWT token service and mocks
│   ├── mocks
│   │   └── token.go
│   ├── README.md
│   ├── token.go
│   └── token_test.go
├── coverage.out
├── coverage.txt
├── database # SQL queries, schemas (SQLC), seeders
│   ├── connect.go
│   ├── query
│   │   ├── card.sql
│   │   ├── merchant_document.sql
│   │   ├── merchant.sql
│   │   ├── README.md
│   │   ├── refresh_token.sql
│   │   ├── reset_token.sql
│   │   ├── role.sql
│   │   ├── saldo.sql
│   │   ├── topup.sql
│   │   ├── transaction.sql
│   │   ├── transfer.sql
│   │   ├── user_role.sql
│   │   ├── user.sql
│   │   └── withdraw.sql
│   ├── README.md
│   ├── schema
│   │   ├── card.sql.go
│   │   ├── db.go
│   │   ├── merchant_document.sql.go
│   │   ├── merchant.sql.go
│   │   ├── models.go
│   │   ├── querier.go
│   │   ├── README.md
│   │   ├── refresh_token.sql.go
│   │   ├── reset_token.sql.go
│   │   ├── role.sql.go
│   │   ├── saldo.sql.go
│   │   ├── topup.sql.go
│   │   ├── transaction.sql.go
│   │   ├── transfer.sql.go
│   │   ├── user_role.sql.go
│   │   ├── user.sql.go
│   │   └── withdraw.sql.go
│   └── seeder 
│       ├── card.go
│       ├── merchant.go
│       ├── README.md
│       ├── role.go
│       ├── saldo.go
│       ├── seed.go
│       ├── topup.go
│       ├── transaction.go
│       ├── transfer.go
│       ├── user.go
│       └── withdraw.go
├── date  # Date parsing and formatting utilities
│   ├── date.go
│   ├── date_test.go
│   └── README.md
├── dotenv  # Environment variable loader
│   ├── dotenv.go
│   └── README.md
├── email  # Email template
│   ├── email.go 
│   └── README.md
├── go.mod
├── go.sum
├── hash  # Password hashing and comparison (bcrypt)
│   ├── hash.go
│   ├── hash_test.go
│   ├── mocks
│   │   └── hash.go
│   └── README.md
├── kafka # Kafka producer/consumer wrappers
│   ├── kafka.go
│   ├── kafka_mocks.go
│   ├── kafka_test.go
│   └── README.md
├── LICENSE
├── logger  # Zap-based logging with mock support
│   ├── logger.go
│   ├── logger_test.go
│   ├── logs
│   │   └── testservice.log
│   ├── mocks
│   │   └── logger.go
│   └── README.md
├── Makefile
├── method_topup # Top-up method validate
│   ├── method.go
│   ├── method_test.go
│   └── README.md
├── otel # OpenTelemetry observability tools
│   ├── otel.go
│   ├── otel_test.go
│   └── README.md
├── random_string # Random string generator
│   ├── random_string.go
│   ├── random_string_test.go
│   └── README.md
├── randomvcc # Random virtual card number generator
│   ├── random.go 
│   ├── random_test.go
│   └── README.md
├── README.md
├── rupiah # Rupiah currency formatter
│   ├── README.md
│   ├── rupiah.go
│   └── rupiah_test.go
└── trace_unic # Unique transaction code tracer
    ├── README.md
    ├── trace_kode_unik.go
    └── trace_kode_unik_test.go
```


## Purpose

The `pkg/` directory serves as a central location for common components and utilities used across the system.
It provides a structured and organized way to group related code, promoting code reusability and maintainability.