package eventsourcing

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event
type Event struct {
	ID            string                 `json:"id"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	EventType     string                 `json:"event_type"`
	Version       int                    `json:"version"`
	Timestamp     time.Time              `json:"timestamp"`
	Data          map[string]interface{} `json:"data"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// NewEvent creates a new event
func NewEvent(aggregateID, aggregateType, eventType string, data map[string]interface{}) *Event {
	return &Event{
		ID:            uuid.New().String(),
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		EventType:     eventType,
		Timestamp:     time.Now(),
		Data:          data,
		Metadata:      make(map[string]interface{}),
	}
}

// EventStore interface for storing and retrieving events
type EventStore interface {
	// Save saves events for an aggregate
	Save(ctx context.Context, events []*Event) error

	// Load loads all events for an aggregate
	Load(ctx context.Context, aggregateID string) ([]*Event, error)

	// LoadFromVersion loads events from a specific version
	LoadFromVersion(ctx context.Context, aggregateID string, version int) ([]*Event, error)
}

// InMemoryEventStore implements EventStore in memory (for development)
type InMemoryEventStore struct {
	events map[string][]*Event
}

// NewInMemoryEventStore creates an in-memory event store
func NewInMemoryEventStore() *InMemoryEventStore {
	return &InMemoryEventStore{
		events: make(map[string][]*Event),
	}
}

func (s *InMemoryEventStore) Save(ctx context.Context, events []*Event) error {
	for _, event := range events {
		s.events[event.AggregateID] = append(s.events[event.AggregateID], event)
	}
	return nil
}

func (s *InMemoryEventStore) Load(ctx context.Context, aggregateID string) ([]*Event, error) {
	return s.events[aggregateID], nil
}

func (s *InMemoryEventStore) LoadFromVersion(ctx context.Context, aggregateID string, version int) ([]*Event, error) {
	allEvents := s.events[aggregateID]
	var filtered []*Event
	for _, e := range allEvents {
		if e.Version >= version {
			filtered = append(filtered, e)
		}
	}
	return filtered, nil
}

// Aggregate represents an event-sourced aggregate root
type Aggregate interface {
	AggregateID() string
	AggregateType() string
	Version() int
	ApplyEvent(event *Event)
	UncommittedEvents() []*Event
	ClearUncommittedEvents()
}

// BaseAggregate provides common aggregate functionality
type BaseAggregate struct {
	id                string
	version           int
	uncommittedEvents []*Event
}

func (a *BaseAggregate) AggregateID() string         { return a.id }
func (a *BaseAggregate) Version() int                { return a.version }
func (a *BaseAggregate) UncommittedEvents() []*Event { return a.uncommittedEvents }
func (a *BaseAggregate) ClearUncommittedEvents()     { a.uncommittedEvents = nil }

func (a *BaseAggregate) RaiseEvent(aggregateType, eventType string, data map[string]interface{}) {
	a.version++
	event := NewEvent(a.id, aggregateType, eventType, data)
	event.Version = a.version
	a.uncommittedEvents = append(a.uncommittedEvents, event)
}

// Example: Account Aggregate

// AccountAggregate represents an account using event sourcing
type AccountAggregate struct {
	BaseAggregate
	OwnerID     string
	AccountType string
	Currency    string
	Balance     float64
	Status      string
}

func NewAccountAggregate(id string) *AccountAggregate {
	return &AccountAggregate{
		BaseAggregate: BaseAggregate{id: id},
	}
}

func (a *AccountAggregate) AggregateType() string { return "Account" }

func (a *AccountAggregate) ApplyEvent(event *Event) {
	switch event.EventType {
	case "AccountCreated":
		a.OwnerID = event.Data["owner_id"].(string)
		a.AccountType = event.Data["account_type"].(string)
		a.Currency = event.Data["currency"].(string)
		a.Balance = 0
		a.Status = "ACTIVE"

	case "MoneyDeposited":
		a.Balance += event.Data["amount"].(float64)

	case "MoneyWithdrawn":
		a.Balance -= event.Data["amount"].(float64)

	case "AccountClosed":
		a.Status = "CLOSED"
	}
	a.version = event.Version
}

// CreateAccount creates a new account
func (a *AccountAggregate) CreateAccount(ownerID, accountType, currency string) {
	a.RaiseEvent("Account", "AccountCreated", map[string]interface{}{
		"owner_id":     ownerID,
		"account_type": accountType,
		"currency":     currency,
	})
}

// Deposit deposits money into the account
func (a *AccountAggregate) Deposit(amount float64, description string) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	a.RaiseEvent("Account", "MoneyDeposited", map[string]interface{}{
		"amount":      amount,
		"description": description,
	})
	return nil
}

// Withdraw withdraws money from the account
func (a *AccountAggregate) Withdraw(amount float64, description string) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if a.Balance < amount {
		return ErrInsufficientFunds
	}
	a.RaiseEvent("Account", "MoneyWithdrawn", map[string]interface{}{
		"amount":      amount,
		"description": description,
	})
	return nil
}

// EventSerializer serializes/deserializes events
type EventSerializer struct{}

func (s *EventSerializer) Serialize(event *Event) ([]byte, error) {
	return json.Marshal(event)
}

func (s *EventSerializer) Deserialize(data []byte) (*Event, error) {
	var event Event
	err := json.Unmarshal(data, &event)
	return &event, err
}

// Errors
var (
	ErrInvalidAmount     = errorf("invalid amount")
	ErrInsufficientFunds = errorf("insufficient funds")
)

type esError string

func (e esError) Error() string { return string(e) }
func errorf(s string) error     { return esError(s) }
