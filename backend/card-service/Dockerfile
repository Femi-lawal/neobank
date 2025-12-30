# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.work go.work.sum ./
COPY shared-lib/go.mod shared-lib/go.sum ./shared-lib/
COPY card-service/go.mod card-service/go.sum ./card-service/

RUN cd shared-lib && go mod download
RUN cd card-service && go mod download

COPY shared-lib/ ./shared-lib/
COPY card-service/ ./card-service/

RUN cd card-service && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/card-service ./cmd/main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/bin/card-service .
# Copy config (optional - file may not exist)
# COPY card-service/config.yaml ./config.yaml

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8085

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8085/health || exit 1

CMD ["./card-service"]
