# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.work go.work.sum ./
COPY shared-lib/go.mod shared-lib/go.sum ./shared-lib/
COPY payment-service/go.mod payment-service/go.sum ./payment-service/

RUN cd shared-lib && go mod download
RUN cd payment-service && go mod download

COPY shared-lib/ ./shared-lib/
COPY payment-service/ ./payment-service/

RUN cd payment-service && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/payment-service ./cmd/main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/bin/payment-service .
COPY payment-service/config.yaml ./config.yaml 2>/dev/null || true

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8083

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8083/health || exit 1

CMD ["./payment-service"]
