# NeoBank Database Migrations

This directory contains SQL migrations for the NeoBank database schema.

## Migration Tool

We recommend using [golang-migrate](https://github.com/golang-migrate/migrate) for managing migrations.

### Installation

```bash
# macOS
brew install golang-migrate

# Windows (scoop)
scoop install migrate

# Go install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Migration Files

| Version | Name | Description |
|---------|------|-------------|
| 000001 | create_users | Users table for identity |
| 000002 | create_accounts | Accounts table for ledger |
| 000003 | create_payments | Payments table for transfers |
| 000004 | create_cards | Cards table for card service |

## Usage

### Apply Migrations

```bash
migrate -path ./migrations -database "postgresql://user:password@localhost:5432/newbank_core?sslmode=disable" up
```

### Rollback Last Migration

```bash
migrate -path ./migrations -database "postgresql://user:password@localhost:5432/newbank_core?sslmode=disable" down 1
```

### Check Current Version

```bash
migrate -path ./migrations -database "postgresql://user:password@localhost:5432/newbank_core?sslmode=disable" version
```

### Create New Migration

```bash
migrate create -ext sql -dir ./migrations -seq create_new_table
```
