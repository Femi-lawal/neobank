# Observability Architecture

NeoBank uses a comprehensive observability stack based on OpenTelemetry (OTEL).

## Telemetry Flow
Services emit Traces, Metrics, and Logs which are aggregated and visualized.

```mermaid
graph LR
    subgraph Services
    ID[Identity Service]
    LG[Ledger Service]
    PY[Payment Service]
    CD[Card Service]
    PD[Product Service]
    end
    
    subgraph Collector
    OTEL[OTEL Collector]
    end
    
    subgraph Backends
    JAE[Jaeger (Tracing)]
    PROM[Prometheus (Metrics)]
    end
    
    subgraph Visualization
    GRAF[Grafana]
    end

    ID -->|OTLP gRPC| OTEL
    LG -->|OTLP gRPC| OTEL
    PY -->|OTLP gRPC| OTEL
    CD -->|OTLP gRPC| OTEL
    PD -->|OTLP gRPC| OTEL

    OTEL -->|Traces| JAE
    OTEL -->|Metrics| PROM

    GRAF -->|Reads| PROM
    GRAF -->|Reads| JAE
```

## Key Components
1.  **OpenTelemetry SDK**: Embedded in Go services (Gin Middleware) to auto-instrument HTTP requests and DB calls.
2.  **OTEL Collector**: Central aggregator. Receives data from all services and exports to backends.
3.  **Jaeger**: Distributed Tracing. Allows visualizing the full path of a request (Frontend -> Payment -> Ledger -> DB).
4.  **Prometheus**: Time-series database for metrics (Request rate, Latency, Errors).
5.  **Grafana**: Unified dashboarding.

## Tracing Strategy
*   **Trace Context Propagation**: `traceparent` headers are passed between microservices.
*   **Spans**: Each major operation (SQL Query, Redis, HTTP Call) creates a Span.
*   **Attributes**: Spans are enriched with `user_id`, `http.status_code`, etc.
