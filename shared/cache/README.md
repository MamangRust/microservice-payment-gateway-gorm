# Distributed Caching and Redis Cluster Strategy

This document details the high-performance caching infrastructure and data distribution strategies implemented via Redis Cluster.

## Redis Cluster Topology

The system utilizes a 6-node Redis Cluster (3 Masters, 3 Replicas) to provide horizonal scalability and high availability.

### Sharding and Data Distribution

- Shard Management: Redis Cluster automatically partitions data across 16,384 hash slots.
- Multi-Node Operations: The application handles multi-key operations using hash tags `{tag}` to ensure keys reside on the same shard where necessary.
- Failover: Each master node is mirrored by a replica. In the event of a master failure, the cluster promotes the replica to master status automatically.

## Cache Use-Cases

1. Identity and Session Management
   Authentication tokens and RBAC snapshots are cached in the Redis Cluster to minimize latency during the gRPC interceptor lifecycle.

2. Behavioral Feature Store (AI Security)
   User activity windowing and behavioral features are stored in Redis to provide sub-millisecond data access for the AI Security engine's risk assessment hooks.

3. Distributed Locking
   Mutually exclusive locks for critical transactional sections are managed using Redis-based primitives to ensure consistency across microservice instances.

## Cluster Management and Observability

The cluster health is continuously monitored using the standard Redis Cluster protocol and integrated into the Unified OTLP pipeline.

### Operational Commands

To check the cluster status within the local environment:

```bash
docker exec -it redis-node-1 redis-cli -a dragon_knight -c cluster nodes
```

Key performance indicators tracked:
- Memory fragmentation ratio.
- Cache hit/miss ratio (monitored via Prometheus).
- Command latency percentiles.
- Shard synchronization lag.
