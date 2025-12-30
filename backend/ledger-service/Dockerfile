# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Disable Go workspace for isolated service build
ENV GOWORK=off

# Copy go module files (without go.work to avoid cross-service dependencies)
COPY shared-lib/go.mod shared-lib/go.sum ./shared-lib/
COPY ledger-service/go.mod ledger-service/go.sum ./ledger-service/

# Download dependencies
RUN cd shared-lib && go mod download
RUN cd ledger-service && go mod download -x

# Copy source code
COPY shared-lib/ ./shared-lib/
COPY ledger-service/ ./ledger-service/

# Build the application with -mod=mod to handle local module replacements
RUN cd ledger-service && CGO_ENABLED=0 GOOS=linux go build -mod=mod -a -installsuffix cgo -o /app/bin/ledger-service ./cmd/main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy binary from builder
COPY --from=builder /app/bin/ledger-service .

# Copy config (if exists)
# Copy config (optional - file may not exist)
# COPY ledger-service/config.yaml ./config.yaml

# Change ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8082

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# Run the application
CMD ["./ledger-service"]
