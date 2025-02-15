# NeoBank Backend Makefile
# Usage: make <target>

.PHONY: all build run test clean help

# Default target
all: help

# =============================================================================
# Development
# =============================================================================

## run-all: Run all services (requires multiple terminals - use docker-compose instead)
run-all:
	@echo "Use 'docker-compose up' for running all services together"
	@echo "Or run each service separately with: make run-<service>"

## run-identity: Run identity service
run-identity:
	cd identity-service && go run ./cmd/main.go

## run-ledger: Run ledger service
run-ledger:
	cd ledger-service && go run ./cmd/main.go

## run-payment: Run payment service
run-payment:
	cd payment-service && go run ./cmd/main.go

## run-product: Run product service
run-product:
	cd product-service && go run ./cmd/main.go

## run-card: Run card service
run-card:
	cd card-service && go run ./cmd/main.go

# =============================================================================
# Build
# =============================================================================

## build-all: Build all services
build-all: build-identity build-ledger build-payment build-product build-card

## build-identity: Build identity service
build-identity:
	cd identity-service && go build -o bin/identity-service ./cmd/main.go

## build-ledger: Build ledger service
build-ledger:
	cd ledger-service && go build -o bin/ledger-service ./cmd/main.go

## build-payment: Build payment service
build-payment:
	cd payment-service && go build -o bin/payment-service ./cmd/main.go

## build-product: Build product service
build-product:
	cd product-service && go build -o bin/product-service ./cmd/main.go

## build-card: Build card service
build-card:
	cd card-service && go build -o bin/card-service ./cmd/main.go

# =============================================================================
# Testing
# =============================================================================

## test-all: Run all tests
test-all:
	go test ./... -v

## test-coverage: Run tests with coverage
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test-identity: Run identity service tests
test-identity:
	cd identity-service && go test ./... -v

## test-ledger: Run ledger service tests
test-ledger:
	cd ledger-service && go test ./... -v

## test-payment: Run payment service tests
test-payment:
	cd payment-service && go test ./... -v

## test-product: Run product service tests
test-product:
	cd product-service && go test ./... -v

## test-card: Run card service tests
test-card:
	cd card-service && go test ./... -v

# =============================================================================
# Dependencies
# =============================================================================

## deps: Download and tidy dependencies
deps:
	go work sync
	cd shared-lib && go mod tidy
	cd identity-service && go mod tidy
	cd ledger-service && go mod tidy
	cd payment-service && go mod tidy
	cd product-service && go mod tidy
	cd card-service && go mod tidy

## deps-update: Update all dependencies
deps-update:
	cd shared-lib && go get -u ./... && go mod tidy
	cd identity-service && go get -u ./... && go mod tidy
	cd ledger-service && go get -u ./... && go mod tidy
	cd payment-service && go get -u ./... && go mod tidy
	cd product-service && go get -u ./... && go mod tidy
	cd card-service && go get -u ./... && go mod tidy

# =============================================================================
# Linting & Formatting
# =============================================================================

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format code
fmt:
	go fmt ./...
	goimports -w .

## vet: Run go vet
vet:
	go vet ./...

# =============================================================================
# Configuration
# =============================================================================

## config-init: Copy example configs to actual configs
config-init:
	@for svc in identity-service ledger-service payment-service product-service card-service; do \
		if [ ! -f $$svc/config.yaml ]; then \
			cp $$svc/config.example.yaml $$svc/config.yaml 2>/dev/null || true; \
			echo "Created $$svc/config.yaml"; \
		else \
			echo "$$svc/config.yaml already exists, skipping"; \
		fi \
	done
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env"; \
	fi

# =============================================================================
# Docker
# =============================================================================

## docker-build: Build all Docker images
docker-build:
	docker build -t neobank/identity-service:latest ./identity-service
	docker build -t neobank/ledger-service:latest ./ledger-service
	docker build -t neobank/payment-service:latest ./payment-service
	docker build -t neobank/product-service:latest ./product-service
	docker build -t neobank/card-service:latest ./card-service

# =============================================================================
# Database
# =============================================================================

## db-migrate: Run database migrations
db-migrate:
	@echo "Migrations are handled by GORM AutoMigrate on service startup"

## db-seed: Seed the database
db-seed:
	psql -h localhost -p 5433 -U user -d newbank_core -f ../infra/seed.sql

# =============================================================================
# Cleanup
# =============================================================================

## clean: Remove build artifacts
clean:
	rm -rf */bin
	rm -f coverage.out coverage.html
	go clean -cache

# =============================================================================
# Help
# =============================================================================

## help: Show this help message
help:
	@echo "NeoBank Backend - Available Commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
