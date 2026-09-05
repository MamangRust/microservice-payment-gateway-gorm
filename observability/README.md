# Unified Observability and Telemetry Pipeline

This document details the telemetry architecture, instrumentation standards, and monitoring infrastructure used across the Distributed Microservice Payment Gateway.

## Observability Architecture

The system implements a standardized OpenTelemetry-native observation model. Every service (Go and Python) is instrumented to emit signals using the OTLP (OpenTelemetry Protocol).

### Unified Signal Routing

1. Collection Layer: The OpenTelemetry (OTEL) Collector serves as the central ingestion point for all logs, metrics, and traces.
2. Processing: Signals are processed, sampled, and enriched with infrastructure metadata (Kubernetes pod IDs, container names, environment tags).
3. Exporting:
   - Metrics -> Prometheus
   - Logs -> Grafana Loki
   - Traces -> Jaeger / Grafana Tempo

## Instrumentation Standards

### Structured Logging
All logs are produced in a machine-readable format (JSON) and include trace context (TraceID, SpanID) to facilitate correlation.
- Log Aggregator: Grafana Loki.
- Query Language: LogQL.

### Distributed Tracing
End-to-end request visibility is maintained using context propagation across gRPC and Kafka boundaries.
- Protocol: W3C Trace Context.
- Visualization: Integrated Tracing dashboards.

### Performance Metrics
System and application-level metrics are exported as Prometheus gauges and counters.
- Core Metrics: RED (Rate, Errors, Duration) pattern.
- Infrastructure Metrics: Host resource utilization (CPU, Memory, Disk, Network).

## Monitoring Infrastructure Components

- Prometheus: Specialized time-series database for metric storage and alerting.
- Grafana Loki: High-performance log aggregation with structured metadata.
- Alertmanager: Orchestrates alerting notifications based on Prometheus rule evaluations.
- Grafana: Unified visualization layer for all telemetry signals.

## Operational Controls

To verify the telemetry pipeline health within the local orchestration:

```bash
# Check OTEL Collector connectivity
curl -v http://localhost:4318/v1/traces

# Verify Prometheus targets
curl http://localhost:9090/api/v1/targets
```