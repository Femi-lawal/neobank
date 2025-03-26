# NeoBank System Design

This document provides an overview of the NeoBank system architecture based on industry best practices from system design principles.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                           CLIENTS                                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐            │
│  │   Web    │  │  Mobile  │  │   API    │  │  Admin   │            │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘            │
└───────┼─────────────┼─────────────┼─────────────┼───────────────────┘
        │             │             │             │
        └─────────────┴──────┬──────┴─────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────────┐
│                            ▼                                         │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                        CDN / WAF                             │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                            │                                         │
│  ┌─────────────────────────┼───────────────────────────────────┐   │
│  │                    Load Balancer                              │   │
│  │               (NGINX / AWS ALB)                               │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                            │                                         │
│  ┌─────────────────────────┼───────────────────────────────────┐   │
│  │                    API Gateway                                │   │
│  │    (Rate Limiting, Auth, Routing, Request Validation)        │   │
│  └─────────────────────────────────────────────────────────────┘   │
└────────────────────────────┼────────────────────────────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────────┐
│                    MICROSERVICES                                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐            │
│  │ Identity │  │  Ledger  │  │ Payment  │  │   Card   │            │
│  │ Service  │  │ Service  │  │ Service  │  │ Service  │            │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘            │
│       │             │             │             │                    │
│       └─────────────┼─────────────┼─────────────┘                    │
│                     │             │                                   │
│  ┌──────────────────┼─────────────┼──────────────────────────────┐  │
│  │              Service Mesh (Istio)                              │  │
│  │     (mTLS, Tracing, Circuit Breaking, Load Balancing)         │  │
│  └────────────────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────────┐
│                    DATA LAYER                                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │  PostgreSQL  │  │    Redis     │  │    Kafka     │              │
│  │   (Primary)  │  │   (Cache)    │  │  (Events)    │              │
│  └──────┬───────┘  └──────────────┘  └──────────────┘              │
│         │                                                            │
│  ┌──────┴───────┐                                                   │
│  │  PostgreSQL  │                                                   │
│  │  (Replica)   │                                                   │
│  └──────────────┘                                                   │
└─────────────────────────────────────────────────────────────────────┘
```

## Design Patterns Implemented

### 1. Microservices Architecture

| Service | Responsibility | Pattern |
|---------|----------------|---------|
| Identity | Auth, Users | API Gateway Pattern |
| Ledger | Accounts, Transactions | Event Sourcing |
| Payment | Transfers | Saga Pattern |
| Card | Card Management | CQRS |
| Product | Catalog | Repository Pattern |

### 2. CQRS (Command Query Responsibility Segregation)

Commands (writes) and queries (reads) are separated for optimization.

```go
// Write path
commandBus.Dispatch(ctx, TransferMoneyCommand{...})

// Read path (optimized)
queryBus.Execute(ctx, GetTransactionsQuery{...})
```

**Location**: `backend/shared-lib/pkg/cqrs/`

### 3. Event Sourcing

State is derived from a sequence of events, enabling:
- Complete audit trail
- Point-in-time recovery
- Event replay

**Location**: `backend/shared-lib/pkg/eventsourcing/`

### 4. Circuit Breaker

Prevents cascading failures when downstream services fail.

```go
cb := resilience.NewCircuitBreaker(config)
err := cb.Execute(func() error {
    return callExternalService()
})
```

**Location**: `backend/shared-lib/pkg/resilience/`

### 5. Service Discovery

Services register themselves and discover others dynamically.

**Location**: `backend/shared-lib/pkg/discovery/`

## Scalability Strategies

### Horizontal Scaling

| Component | Strategy | Implementation |
|-----------|----------|----------------|
| Services | Kubernetes HPA | Auto-scale on CPU/memory |
| Database | Read replicas | PostgreSQL streaming replication |
| Cache | Redis Cluster | Sharded across nodes |

### Caching Strategy

| Layer | Cache | TTL | Purpose |
|-------|-------|-----|---------|
| CDN | CloudFront | 24h | Static assets |
| Application | Redis | 5-60min | Session, API responses |
| Database | Connection pool | N/A | Query caching |

### Database Sharding (Future)

| Shard Key | Services | Strategy |
|-----------|----------|----------|
| User ID | Identity, Card | Hash-based |
| Account ID | Ledger, Payment | Range-based |

## Reliability Patterns

### 1. Rate Limiting

- Per-IP: 100 req/min
- Per-User: 1000 req/min
- Sensitive endpoints: 10 req/min

### 2. Retry with Exponential Backoff

```go
for attempt := 0; attempt < maxRetries; attempt++ {
    if err := operation(); err == nil {
        break
    }
    time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
}
```

### 3. Saga Pattern for Distributed Transactions

```
Transfer Saga:
1. Debit source account (with compensation: credit back)
2. Credit destination account (with compensation: debit back)
3. Record transaction
```

## Data Consistency

| Requirement | Approach |
|-------------|----------|
| Account balance | Strong consistency (ACID) |
| Transaction history | Eventual consistency (Event Sourcing) |
| User sessions | Eventual consistency (Redis) |
| Product catalog | Read-your-writes consistency |

## Security Layers

1. **Network**: WAF, DDoS protection
2. **Transport**: TLS 1.3, mTLS between services
3. **Application**: JWT, input validation
4. **Data**: Encryption at rest, PII masking

## Observability Stack

| Pillar | Tool | Purpose |
|--------|------|---------|
| Metrics | Prometheus + Grafana | Performance monitoring |
| Logs | Loki | Log aggregation |
| Traces | Jaeger | Distributed tracing |
| APM | OpenTelemetry | Application performance |

## References

- [System Design Primer](https://github.com/donnemartin/system-design-primer)
- [Karan's System Design](https://github.com/karanpratapsingh/system-design)
