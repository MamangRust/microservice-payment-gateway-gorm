# Protocol System and gRPC Specifications

This document details the interface definitions and service mesh communication protocols used across the Distributed Microservice Payment Gateway.

## Protocol Buffers v3

The system utilizes Protocol Buffers (proto3) as the authoritative language for defining service interfaces and data schemas. This ensures strongly-typed, binary-efficient, and cross-language compatibility.

### Core Service Definitions

1. Authentication (auth.proto)
   - Procedures for identity verification and token lifecycle management.
   - Fault-tolerant RBAC check implementations.

2. Transaction Handling (transaction.proto)
   - Definitions for financial event propagation.
   - Synchronous hooks for state transition authorization.

3. AI Security (ai_security/ai_security.proto)
   - Real-time fraud detection hooks used for pre-transaction risk assessment.

## Service Mesh Coordination

All inter-service communication is conducted via gRPC over HTTP/2.

- Load Balancing: Managed via internal Kubernetes service discovery.
- Health Probes: Standardized gRPC health check protocols implemented across all services.
- Deadline Propagation: Contextual deadlines are strictly enforced to prevent cascading failure scenarios.

## Development and Generation

To regenerate the Go bindings for any proto definition, use the following command from the project root:

```bash
just generate-proto
```

This will output the generated code into the `pb/` directory, structured by domain.
