# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go workspace files
COPY go.work go.work.sum ./
COPY shared-lib/go.mod shared-lib/go.sum ./shared-lib/
COPY identity-service/go.mod identity-service/go.sum ./identity-service/

# Download dependencies
RUN cd shared-lib && go mod download
RUN cd identity-service && go mod download

# Copy source code
COPY shared-lib/ ./shared-lib/
COPY identity-service/ ./identity-service/

# Build the application
RUN cd identity-service && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/identity-service ./cmd/main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy binary from builder
COPY --from=builder /app/bin/identity-service .

# Copy config (optional - file may not exist in all environments)
# Note: This COPY is commented out as config.yaml is optional for demo
# COPY identity-service/config.yaml ./config.yaml

# Change ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Run the application
CMD ["./identity-service"]
