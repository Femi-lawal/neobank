# NeoBank Disaster Recovery Plan

## Overview

This document outlines the disaster recovery (DR) procedures for NeoBank production systems.

## Recovery Objectives

| Metric | Target | Maximum Tolerable |
|--------|--------|-------------------|
| **RTO** (Recovery Time Objective) | 1 hour | 4 hours |
| **RPO** (Recovery Point Objective) | 5 minutes | 1 hour |

## Disaster Scenarios

### Tier 1: Complete Region Failure

**Scenario**: Primary cloud region becomes unavailable.

**Response**:
1. Traffic automatically routes to secondary region via DNS failover
2. Secondary databases promoted to primary
3. Notify customers via status page

**Recovery Steps**:
```bash
# 1. Verify secondary region health
kubectl --context=dr-region get pods -n neobank

# 2. Promote read replicas to primary
kubectl apply -f k8s/dr/promote-replicas.yaml

# 3. Update DNS to point to DR region
./scripts/dns-failover.sh activate-dr

# 4. Verify services
./scripts/smoke-test.sh --region=dr
```

### Tier 2: Database Corruption/Loss

**Scenario**: Primary database becomes corrupted or unavailable.

**Response**:
1. Failover to read replica
2. Restore from latest backup if needed
3. Replay transaction logs

**Recovery Steps**:
```bash
# 1. Switch to replica
kubectl patch service postgres-primary -p '{"spec":{"selector":{"role":"replica"}}}'

# 2. If restore needed
./scripts/db-restore.sh --point-in-time="2025-03-21T10:00:00Z"

# 3. Verify data integrity
./scripts/verify-db-integrity.sh
```

### Tier 3: Service Degradation

**Scenario**: Individual service or component failure.

**Response**:
1. Circuit breakers activate
2. Auto-scaling increases replicas
3. On-call responds

**Recovery Steps**:
```bash
# 1. Check service health
kubectl get pods -l app=failing-service -n neobank

# 2. Restart unhealthy pods
kubectl rollout restart deployment/failing-service -n neobank

# 3. Scale up if needed
kubectl scale deployment/failing-service --replicas=5 -n neobank
```

## Backup Strategy

### Database Backups

| Type | Frequency | Retention | Location |
|------|-----------|-----------|----------|
| Full backup | Daily | 30 days | S3 (cross-region) |
| Incremental | Hourly | 7 days | S3 |
| WAL logs | Continuous | 24 hours | S3 + local |
| Point-in-time | Continuous | 7 days | RDS automated |

### Application State

| Component | Backup Method | Frequency |
|-----------|---------------|-----------|
| Kubernetes configs | GitOps (ArgoCD) | Real-time |
| Secrets | Vault snapshots | Daily |
| Redis cache | AOF persistence | Continuous |

## Runbooks

### Pre-Incident Preparation

1. ✅ Automated backups verified weekly
2. ✅ DR region tested monthly
3. ✅ Runbooks reviewed quarterly
4. ✅ On-call rotation maintained

### During Incident

1. **Assemble team**: Page on-call, assemble incident team
2. **Communicate**: Update status page, notify stakeholders
3. **Diagnose**: Identify root cause using observability tools
4. **Recover**: Execute appropriate runbook
5. **Verify**: Confirm services are healthy

### Post-Incident

1. Write incident report within 24 hours
2. Conduct blameless postmortem within 1 week
3. Implement preventive measures within 2 weeks
4. Update runbooks if needed

## Communication Plan

### Internal

| Event | Channel | Audience |
|-------|---------|----------|
| Incident declared | Slack #incident | Engineering |
| Updates (every 30 min) | Slack #incident | Engineering |
| Resolution | Email | All staff |

### External

| Event | Channel | Audience |
|-------|---------|----------|
| Outage detected | Status page | Customers |
| Updates (every 1 hr) | Status page | Customers |
| Resolution | Status page + Email | Customers |

## DR Testing Schedule

| Test Type | Frequency | Duration |
|-----------|-----------|----------|
| Backup restore | Weekly | 1 hour |
| Failover (non-prod) | Monthly | 2 hours |
| Full DR drill | Quarterly | 4 hours |
| Chaos engineering | Weekly | 1 hour |

## Contacts

| Role | Primary | Secondary |
|------|---------|-----------|
| Incident Commander | Platform Lead | Engineering Manager |
| DB Admin | DBA On-call | Platform Team |
| Communications | Product Manager | Support Lead |

## Tools

- **Monitoring**: Prometheus, Grafana
- **Alerting**: PagerDuty
- **Status Page**: Statuspage.io
- **Runbooks**: Confluence/GitHub Wiki
- **Communication**: Slack, Zoom
