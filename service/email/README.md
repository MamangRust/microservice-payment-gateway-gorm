# Payment Gateway Email Service

## Overview

`EmailService` is a microservice designed to handle **asynchronous email delivery** triggered by events from other services via **Kafka**. It is responsible for processing events such as OTP codes, registration confirmations, transaction notifications, and merchant status updates, and then sending the corresponding emails via SMTP.


This service is event-driven, meaning it reacts to messages produced to specific Kafka topics, decoupling it from the services that generate those events.


### 📊 Monitoring & Observability

#### Prometheus Metrics
  - email_service_requests_total{method, status}
    → Total number of email requests by method and response status.
  - email_service_request_duration_seconds{method}
    → Histogram tracking how long each email sending process takes.

#### OpenTelemetry Tracing
Each service is assigned a tracer for distributed tracing:
- `email-service`



#### 📬 Kafka Topics

| Topic Name                                     | Purpose                                           |
|------------------------------------------------|---------------------------------------------------|
| `email-service-topic-auth-register`            | Send welcome email after user registration        |
| `email-service-topic-auth-forgot-password`     | Send OTP code for password reset                  |
| `email-service-topic-auth-verify-code-success` | Notify successful OTP verification                |
| `email-service-topic-saldo-create`             | Notify when new balance is created                |
| `email-service-topic-topup-create`             | Notify user of a successful top-up                |
| `email-service-topic-transaction-create`       | Notify user of a successful transaction           |
| `email-service-topic-transfer-create`          | Notify user of successful transfer                |
| `email-service-topic-withdraw-create`          | Notify user of successful withdrawal              |
| `email-service-topic-withdraw-update`          | Notify user of withdrawal record updates          |
| `email-service-topic-merchant-create`          | Notify merchant creation success                  |
| `email-service-topic-merchant-update-status`   | Notify merchant status updates                    |
| `email-service-topic-merchant-document-create` | Notify merchant document creation                 |
| `email-service-topic-merchant-document-update-status` | Notify document status updates            |