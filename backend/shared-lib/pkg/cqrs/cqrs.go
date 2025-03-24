package cqrs

import (
	"context"
	"time"
)

// Command represents a write operation that changes system state
type Command interface {
	CommandName() string
	Validate() error
}

// Query represents a read operation that doesn't change state
type Query interface {
	QueryName() string
}

// CommandHandler handles command execution
type CommandHandler interface {
	Handle(ctx context.Context, cmd Command) error
}

// QueryHandler handles query execution
type QueryHandler[T any] interface {
	Handle(ctx context.Context, query Query) (T, error)
}

// CommandBus routes commands to their handlers
type CommandBus struct {
	handlers map[string]CommandHandler
}

// NewCommandBus creates a new command bus
func NewCommandBus() *CommandBus {
	return &CommandBus{
		handlers: make(map[string]CommandHandler),
	}
}

// Register registers a handler for a command type
func (b *CommandBus) Register(commandName string, handler CommandHandler) {
	b.handlers[commandName] = handler
}

// Dispatch dispatches a command to its handler
func (b *CommandBus) Dispatch(ctx context.Context, cmd Command) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	handler, exists := b.handlers[cmd.CommandName()]
	if !exists {
		return ErrHandlerNotFound
	}

	return handler.Handle(ctx, cmd)
}

// QueryBus routes queries to their handlers
type QueryBus struct {
	handlers map[string]interface{}
}

// NewQueryBus creates a new query bus
func NewQueryBus() *QueryBus {
	return &QueryBus{
		handlers: make(map[string]interface{}),
	}
}

// Example Commands

// CreateAccountCommand represents a command to create an account
type CreateAccountCommand struct {
	UserID      string
	AccountType string
	Currency    string
}

func (c CreateAccountCommand) CommandName() string { return "CreateAccount" }
func (c CreateAccountCommand) Validate() error {
	if c.UserID == "" {
		return ErrInvalidUserID
	}
	if c.AccountType == "" {
		return ErrInvalidAccountType
	}
	return nil
}

// TransferMoneyCommand represents a command to transfer money
type TransferMoneyCommand struct {
	FromAccountID string
	ToAccountID   string
	Amount        float64
	Currency      string
	Description   string
}

func (c TransferMoneyCommand) CommandName() string { return "TransferMoney" }
func (c TransferMoneyCommand) Validate() error {
	if c.FromAccountID == "" || c.ToAccountID == "" {
		return ErrInvalidAccountID
	}
	if c.Amount <= 0 {
		return ErrInvalidAmount
	}
	if c.FromAccountID == c.ToAccountID {
		return ErrSameAccount
	}
	return nil
}

// Example Queries

// GetAccountQuery queries an account by ID
type GetAccountQuery struct {
	AccountID string
}

func (q GetAccountQuery) QueryName() string { return "GetAccount" }

// GetTransactionsQuery queries transactions for an account
type GetTransactionsQuery struct {
	AccountID string
	StartDate time.Time
	EndDate   time.Time
	Limit     int
}

func (q GetTransactionsQuery) QueryName() string { return "GetTransactions" }

// Errors
var (
	ErrHandlerNotFound    = errorf("handler not found for command")
	ErrInvalidUserID      = errorf("invalid user ID")
	ErrInvalidAccountType = errorf("invalid account type")
	ErrInvalidAccountID   = errorf("invalid account ID")
	ErrInvalidAmount      = errorf("invalid amount")
	ErrSameAccount        = errorf("cannot transfer to same account")
)

type cqrsError string

func (e cqrsError) Error() string { return string(e) }
func errorf(s string) error       { return cqrsError(s) }
