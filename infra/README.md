# NeoBank Infrastructure

This directory contains the infrastructure configuration for the NeoBank platform.

## 🚀 Quick Start

```bash
# Start all infrastructure services
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f
```

## 📦 Services

| Service | Port | Description |
|---------|------|-------------|
| **PostgreSQL** | 5433 | Primary database |
| **Redis** | 6379 | Caching layer |
| **Kafka** | 9092 | Event streaming |
| **Kafka UI** | 8090 | Kafka management UI |
| **Prometheus** | 9090 | Metrics collection |
| **Grafana** | 3000 | Dashboards & visualization |
| **Jaeger** | 16686 | Distributed tracing |
| **OTel Collector** | 4317/4318 | Telemetry pipeline |

## 📊 Observability Stack

### Accessing Dashboards

| Dashboard | URL | Credentials |
|-----------|-----|-------------|
| Grafana | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | - |
| Jaeger | http://localhost:16686 | - |
| Kafka UI | http://localhost:8090 | - |

### Pre-configured Grafana Dashboards

1. **NeoBank - Services Overview** (`neobank-overview`)
   - Service health status (UP/DOWN)
   - Request rate per service
   - Response time (p95)
   - Payment metrics (transfers, error rate)
   - Infrastructure metrics (Redis, Kafka, Postgres)

### Prometheus Alert Rules

Alerts are defined in `observability/prometheus/alerts/neobank.yml`:

| Alert | Severity | Description |
|-------|----------|-------------|
| ServiceDown | Critical | Any service is unreachable |
| PaymentServiceDown | Critical | Payment service specifically down |
| HighPaymentFailureRate | Critical | >1% payment failures |
| HighErrorRate | Warning | >5% HTTP 5xx errors |
| HighLatency | Warning | p95 latency >1s |
| PostgresDown | Critical | Database unreachable |
| HighDatabaseConnections | Warning | >80% connection usage |
| RedisDown | Critical | Cache unreachable |
| KafkaConsumerLag | Warning | Consumer lag >1000 |

### OpenTelemetry Configuration

The OTEL Collector (`observability/otel/otel-collector-config.yaml`) provides:

- **Receivers**: OTLP (gRPC/HTTP), Prometheus scraping, Host metrics
- **Processors**: Batching, Memory limiting, Tail sampling, Resource attributes
- **Exporters**: Jaeger (traces), Prometheus (metrics)

## 🔧 Configuration

### Directory Structure

```
infra/
├── docker-compose.yml          # Main orchestration file
├── seed.sql                    # Database seed data
└── observability/
    ├── prometheus/
    │   ├── prometheus.yml      # Scrape configuration
    │   └── alerts/
    │       └── neobank.yml     # Alert rules
    ├── grafana/
    │   ├── provisioning/
    │   │   ├── datasources/    # Auto-configure data sources
    │   │   └── dashboards/     # Dashboard provisioning
    │   └── dashboards/
    │       └── neobank-overview.json
    └── otel/
        └── otel-collector-config.yaml
```

### Adding Custom Metrics to Services

To expose Prometheus metrics from your Go services, add the Gin middleware:

```go
import "github.com/penglongli/gin-metrics/ginmetrics"

func main() {
    r := gin.Default()
    
    // Add metrics middleware
    m := ginmetrics.GetMonitor()
    m.SetMetricPath("/metrics")
    m.Use(r)
    
    // ... rest of your routes
}
```

### Sending Traces via OpenTelemetry

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
)

func initTracer() {
    exporter, _ := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint("localhost:4317"),
        otlptracegrpc.WithInsecure(),
    )
    // Configure trace provider...
}
```

## 🗄️ Database

### Seeding the Database

```bash
# Run seed script
psql -h localhost -p 5433 -U user -d newbank_core -f seed.sql

# Or using docker
docker exec -i newbank_postgres psql -U user -d newbank_core < seed.sql
```

### Default Credentials

| Component | Username | Password |
|-----------|----------|----------|
| PostgreSQL | user | password |
| Grafana | admin | admin |

## 🐳 Docker Commands

```bash
# Start specific services only
docker-compose up -d postgres redis kafka

# Restart a specific service
docker-compose restart prometheus

# View resource usage
docker stats

# Clean up volumes (WARNING: deletes data!)
docker-compose down -v
```

## 🔍 Troubleshooting

### Prometheus Not Scraping Services

1. Check that services expose `/metrics` endpoint
2. Verify `host.docker.internal` resolves correctly
3. Check Prometheus targets: http://localhost:9090/targets

### Kafka Consumer Lag

1. Check consumer group status in Kafka UI
2. Verify consumer is connected: `docker logs newbank_kafka`
3. Check topic exists: Kafka UI → Topics

### Grafana Dashboard Not Loading

1. Check datasource connection: Grafana → Configuration → Data Sources
2. Verify Prometheus is running: http://localhost:9090
3. Check provisioning logs: `docker logs newbank_grafana`
