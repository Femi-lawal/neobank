# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Disable Go workspace for isolated service build
ENV GOWORK=off

# Copy go module files (without go.work to avoid cross-service dependencies)
COPY shared-lib/go.mod shared-lib/go.sum ./shared-lib/
COPY product-service/go.mod product-service/go.sum ./product-service/

RUN cd shared-lib && go mod download
RUN cd product-service && go mod download -x

COPY shared-lib/ ./shared-lib/
COPY product-service/ ./product-service/

# Build with -mod=mod to handle local module replacements
RUN cd product-service && CGO_ENABLED=0 GOOS=linux go build -mod=mod -a -installsuffix cgo -o /app/bin/product-service ./cmd/main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/bin/product-service .
# Copy config (optional - file may not exist)
# COPY product-service/config.yaml ./config.yaml

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8084

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8084/health || exit 1

CMD ["./product-service"]
