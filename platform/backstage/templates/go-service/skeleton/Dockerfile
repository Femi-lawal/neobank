FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./cmd/main.go

FROM gcr.io/distroless/static-debian11

COPY --from=builder /service /service

EXPOSE ${{ values.port }}

ENTRYPOINT ["/service"]
