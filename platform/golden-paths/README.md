# NeoBank Golden Paths

**Golden Paths** are the opinionated, supported paths for building services at NeoBank. Following these paths ensures consistency, security, and operational excellence.

## What is a Golden Path?

A Golden Path is the **recommended way** to accomplish a task. It's not the only way, but it's:
- ✅ Tested and proven
- ✅ Fully supported
- ✅ Compliant with security requirements
- ✅ Observable and monitorable

## Service Creation Golden Path

### 1. Use the Backstage Template

```bash
# Via Backstage UI
1. Go to Create → Templates
2. Select "NeoBank Go Microservice"
3. Fill in service details
4. Click Create
```

### 2. Standard Service Structure

```
my-service/
├── cmd/
│   └── main.go           # Application entry point
├── internal/
│   ├── handler/          # HTTP handlers
│   ├── service/          # Business logic
│   ├── repository/       # Data access
│   └── model/            # Domain models
├── api/
│   └── openapi.yaml      # API specification
├── Dockerfile
├── catalog-info.yaml     # Backstage registration
├── go.mod
└── README.md
```

### 3. Required Components

| Component | Tool | Why |
|-----------|------|-----|
| HTTP Framework | Gin | Standard, performance |
| Database | GORM + PostgreSQL | Type safety, migrations |
| Cache | Redis | Performance, sessions |
| Messaging | Kafka | Event-driven, durability |
| Logging | Zap | Structured, fast |
| Config | Viper | 12-factor, env vars |
| Metrics | Prometheus | Observability |
| Tracing | OpenTelemetry | Distributed tracing |

## API Design Golden Path

### 1. RESTful Conventions

```yaml
# Good ✅
GET    /api/v1/accounts
GET    /api/v1/accounts/{id}
POST   /api/v1/accounts
PUT    /api/v1/accounts/{id}
DELETE /api/v1/accounts/{id}

# Bad ❌
GET    /api/v1/getAccounts
POST   /api/v1/createAccount
```

### 2. Response Format

```json
{
  "data": { ... },
  "meta": {
    "page": 1,
    "total": 100
  }
}
```

### 3. Error Format

```json
{
  "error": {
    "code": "INSUFFICIENT_FUNDS",
    "message": "Account balance too low",
    "details": { ... }
  }
}
```

## Deployment Golden Path

### 1. Container Build

```dockerfile
# Multi-stage build (required)
FROM golang:1.21-alpine AS builder
# ... build steps ...

FROM gcr.io/distroless/static-debian11
# Minimal runtime image
```

### 2. Kubernetes Deployment

- Use Helm chart in `/helm/neobank`
- Configure via values.yaml
- Deploy with ArgoCD

### 3. Required Labels

```yaml
metadata:
  labels:
    app: my-service
    app.kubernetes.io/name: my-service
    app.kubernetes.io/part-of: neobank
    app.kubernetes.io/version: "1.0.0"
```

## Security Golden Path

### 1. Authentication
- All endpoints require JWT (except /health)
- Use shared-lib auth middleware

### 2. Input Validation
- Use shared-lib input validation middleware
- Define Zod schemas for all inputs

### 3. Secrets
- Never in code or config files
- Use Kubernetes Secrets or Vault

## Testing Golden Path

| Type | Tool | Coverage |
|------|------|----------|
| Unit | Go testing + testify | 80% |
| Integration | Go testing | Critical paths |
| E2E | Playwright | User flows |
| Load | k6 | Performance gates |

## Observability Golden Path

### 1. Metrics (Required)
- `http_requests_total`
- `http_request_duration_seconds`
- `db_query_duration_seconds`

### 2. Logs (Required)
- Structured JSON format
- Include request_id, user_id
- Level: INFO, WARN, ERROR

### 3. Traces (Required)
- OpenTelemetry instrumentation
- Trace context propagation

## Getting Help

- 📖 **Docs**: Check `/docs` folder
- 💬 **Slack**: #platform-support
- 🎫 **Issues**: GitHub Issues

## Why Follow Golden Paths?

| Benefit | Description |
|---------|-------------|
| **Speed** | Pre-built templates reduce setup time |
| **Security** | Vetted patterns reduce vulnerabilities |
| **Support** | Platform team supports golden paths |
| **Consistency** | Common patterns across all services |
| **Onboarding** | New developers get up to speed faster |
