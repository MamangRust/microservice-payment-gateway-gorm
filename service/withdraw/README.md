# Payment Gateway Withdraw Service

## Overview

`WithdrawService` is a microservice responsible for managing withdrawal transactions. It provides monthly and yearly statistics categorized by status, amount, and card number. The service is designed with high observability using Prometheus, OpenTelemetry, and Zap Logger.


### 🔄 Service Architecture

**WithdrawService** is structured into multiple specialized services:

- **`withdrawCommandService`** — Manages write operations such as create, update, soft-delete, restore, and permanent delete.
- **`withdrawQueryService`** — Handles read operations like listing, searching, and retrieving withdrawals.
- **`withdrawStatisticService`** — Provides general withdrawal statistics across all cards.
- **`withdrawStatisticByCardService`** — Provides withdrawal statistics grouped by card number.

-----

### 📌 Available RPC Methods

#### 📘 Query Service Methods

| Method                         | Description                                      |
|-------------------------------|--------------------------------------------------|
| `FindAllWithdraw`             | Retrieve all withdrawal records with pagination. |
| `FindAllWithdrawByCardNumber` | Retrieve withdrawals filtered by card number.    |
| `FindByIdWithdraw`            | Retrieve withdrawal by ID.                       |
| `FindByCardNumber`            | Retrieve withdrawals by card number.             |
| `FindByActive`                | Get all active (non-deleted) withdrawals.        |
| `FindByTrashed`               | Get all soft-deleted withdrawals.                |

---

#### 📊 Statistic Service Methods


| Method                            | Description                             |
|----------------------------------|-----------------------------------------|
| `FindMonthlyWithdrawStatusSuccess` | Monthly successful withdrawals.       |
| `FindYearlyWithdrawStatusSuccess`  | Yearly successful withdrawals.        |
| `FindMonthlyWithdrawStatusFailed`  | Monthly failed withdrawals.           |
| `FindYearlyWithdrawStatusFailed`   | Yearly failed withdrawals.            |
| `FindMonthlyWithdraws`             | Monthly withdrawal amounts.           |
| `FindYearlyWithdraws`              | Yearly withdrawal amounts.            |
| `FindMonthlyWithdrawStatusSuccessCardNumber` | Monthly successful withdrawals by card number.  |
| `FindYearlyWithdrawStatusSuccessCardNumber`  | Yearly successful withdrawals by card number.   |
| `FindMonthlyWithdrawStatusFailedCardNumber`  | Monthly failed withdrawals by card number.      |
| `FindYearlyWithdrawStatusFailedCardNumber`   | Yearly failed withdrawals by card number.       |
| `FindMonthlyWithdrawsByCardNumber`           | Monthly withdrawal amounts by card number.      |
| `FindYearlyWithdrawsByCardNumber`            | Yearly withdrawal amounts by card number.       |
| `CreateWithdraw`             | Create a new withdrawal record.                 |
| `UpdateWithdraw`             | Update an existing withdrawal.                  |
| `TrashedWithdraw`            | Soft-delete a withdrawal.                       |
| `RestoreWithdraw`            | Restore a soft-deleted withdrawal.              |
| `DeleteWithdrawPermanent`    | Permanently delete a withdrawal.                |
| `RestoreAllWithdraw`         | Restore all soft-deleted withdrawals.           |
| `DeleteAllWithdrawPermanent` | Permanently delete all soft-deleted withdrawals.|

---

## 📊 Monitoring & Observability

#### 📈 Prometheus Metrics

Each service exposes:

- **Request Counter** — Total requests, labeled by method and status.
- **Request Duration Histogram** — Duration metrics per method.

**Metric name examples:**
- `withdraw_command_service_request_total`
- `withdraw_command_service_request_duration_seconds`
- `withdraw_query_service_request_total`
- `withdraw_query_service_request_duration_seconds`
- `withdraw_statistic_service_request_total`
- `withdraw_statistic_service_request_duration_seconds`
- `withdraw_statistic_by_card_service_request_total`
- `withdraw_statistic_by_card_service_request_duration_seconds`


#### 🔍 OpenTelemetry Tracing

Each component uses a dedicated tracer for distributed tracing:

- `withdraw-command-service`
- `withdraw-query-service`
- `withdraw-statistic-service`
- `withdraw-statistic-by-card-service`

----


### 📬 Kafka Integration

The WithdrawService publishes withdrawal events to Kafka to enable asynchronous processing. These events trigger essential downstream operations like email notifications and transaction records.


| Kafka Topic                                            | Purpose                                                               | Consumer Group        |
| ------------------------------------------------------ | --------------------------------------------------------------------- | --------------------- |
| `email-service-topic-withdraw-create`                 | Notify user when a new withdraw is successfully created               | `email-service-group` |
| `email-service-topic-withdraw-update`                  | Notify user when a withdraw record is updated                         | `email-service-group` |
----

### 📏 Daily Withdrawal Limit

Each card has a per-day withdrawal cap, enforced before any balance mutation:

- **Env:** `WITHDRAW_DAILY_LIMIT` (smallest currency unit, default `10_000_000`)
- **Rule:** `sum(today's successful withdrawals) + new amount <= limit`; otherwise the
  request is rejected with `400 Bad Request: Daily withdrawal limit exceeded` and the
  balance is left untouched.
- **Enforced in:** `CreateWithdraw` and the positive-delta path of `UpdateWithdraw`.
- **Query:** `GetTodayWithdrawSumByCardNumber` (`service/withdraw/sqlc/query/withdraw.sql`).

To disable the limit, set `WITHDRAW_DAILY_LIMIT=0`.
----