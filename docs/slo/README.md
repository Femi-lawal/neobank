# NeoBank Service Level Objectives (SLOs)

This document defines the Service Level Objectives for NeoBank services.

## Terminology

| Term | Definition |
|------|------------|
| **SLA** | Service Level Agreement - Contract with customers |
| **SLO** | Service Level Objective - Internal targets |
| **SLI** | Service Level Indicator - Actual measurements |
| **Error Budget** | Acceptable amount of downtime/errors |

## Platform-Wide SLOs

### Availability

| Service | Target | Error Budget (monthly) |
|---------|--------|------------------------|
| Identity Service | 99.95% | 21.6 minutes |
| Ledger Service | 99.99% | 4.3 minutes |
| Payment Service | 99.99% | 4.3 minutes |
| Card Service | 99.9% | 43.2 minutes |
| Product Service | 99.9% | 43.2 minutes |
| Frontend | 99.95% | 21.6 minutes |

### Latency (p99)

| Endpoint Type | Target | Alert Threshold |
|---------------|--------|-----------------|
| GET requests | < 100ms | > 200ms |
| POST requests | < 500ms | > 1s |
| Payment processing | < 2s | > 5s |
| Authentication | < 200ms | > 500ms |
| Dashboard load | < 1s | > 2s |

### Error Rate

| Service | Target | Alert Threshold |
|---------|--------|-----------------|
| All services | < 0.1% | > 1% |
| Payment service | < 0.01% | > 0.1% |

## Service-Specific SLOs

### Identity Service

| SLI | Target |
|-----|--------|
| Login success rate | > 99.9% |
| Token generation time | < 50ms |
| Failed login rate (false positives) | < 0.01% |

### Ledger Service

| SLI | Target |
|-----|--------|
| Transaction recording | 100% durability |
| Balance accuracy | 100% |
| Query response time | < 50ms |

### Payment Service

| SLI | Target |
|-----|--------|
| Transfer success rate | > 99.99% |
| Processing time | < 2s |
| Duplicate detection | 100% |

## SLI Collection

### Metrics Endpoints

```prometheus
# Availability
up{job="neobank-*"}

# Latency
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

# Error Rate
sum(rate(http_requests_total{code=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))
```

### Alerting Rules

```yaml
groups:
  - name: slo-alerts
    rules:
      - alert: HighErrorRate
        expr: sum(rate(http_requests_total{code=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) > 0.01
        for: 5m
        labels:
          severity: critical
          
      - alert: HighLatency
        expr: histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m])) > 0.5
        for: 5m
        labels:
          severity: warning
```

## Error Budget Policy

### When Error Budget is Healthy (> 50%)

- Normal development velocity
- Feature releases continue
- Experimentation allowed

### When Error Budget is Low (< 50%)

- Increased review for changes
- Focus on reliability improvements
- Limit risky deployments

### When Error Budget is Exhausted (< 0%)

- Feature freeze
- All hands on reliability
- Postmortem required for any incident

## Incident Classification

| Severity | Impact | Response Time | Resolution Time |
|----------|--------|---------------|-----------------|
| P1 - Critical | All users affected | 5 min | 1 hour |
| P2 - Major | Many users affected | 15 min | 4 hours |
| P3 - Minor | Some users affected | 1 hour | 24 hours |
| P4 - Low | Few users affected | 4 hours | 1 week |

## Review Schedule

- **Weekly**: SLO dashboard review
- **Monthly**: Error budget review
- **Quarterly**: SLO target review and adjustment
