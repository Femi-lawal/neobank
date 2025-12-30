package service

import (
	"testing"

	"github.com/femi-lawal/new_bank/backend/ledger-service/internal/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockLedgerRepo struct {
	mock.Mock
}

func (m *MockLedgerRepo) CreateAccount(acc *model.Account) error {
	args := m.Called(acc)
	return args.Error(0)
}

func (m *MockLedgerRepo) ListAccounts() ([]model.Account, error) {
	args := m.Called()
	return args.Get(0).([]model.Account), args.Error(1)
}

func (m *MockLedgerRepo) PostTransaction(entry *model.JournalEntry) error {
	args := m.Called(entry)
	return args.Error(0)
}

func (m *MockLedgerRepo) GetAccount(id string) (*model.Account, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Account), args.Error(1)
}

func (m *MockLedgerRepo) ListAccountsByUser(userID string) ([]model.Account, error) {
	args := m.Called(userID)
	return args.Get(0).([]model.Account), args.Error(1)
}

func TestCreateAccount(t *testing.T) {
	mockRepo := new(MockLedgerRepo)
	service := NewLedgerService(mockRepo)

	// Expectation
	mockRepo.On("CreateAccount", mock.AnythingOfType("*model.Account")).Return(nil)

	// Execute
	acc, err := service.CreateAccount(uuid.New().String(), "123", "Checking", "USD", model.Asset)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "123", acc.AccountNumber)
	assert.Equal(t, "Checking", acc.Name)
	assert.True(t, acc.CachedBalance.IsZero())
	mockRepo.AssertExpectations(t)
}

func TestPostTransaction(t *testing.T) {
	mockRepo := new(MockLedgerRepo)
	service := NewLedgerService(mockRepo)

	// 1. Invalid Postings
	_, err := service.PostTransaction("Test", []PostingRequest{})
	assert.Error(t, err)

	// 2. Success
	// This test is tricky because PostTransaction parses UUIDs and Decimals from strings.
	// We need valid UUIDs.
	uuid1 := "00000000-0000-0000-0000-000000000001"
	uuid2 := "00000000-0000-0000-0000-000000000002"

	postings := []PostingRequest{
		{AccountID: uuid1, Amount: "100.00", Direction: 1},
		{AccountID: uuid2, Amount: "100.00", Direction: -1},
	}

	mockRepo.On("PostTransaction", mock.AnythingOfType("*model.JournalEntry")).Return(nil)

	entry, err := service.PostTransaction("Transfer", postings)
	assert.NoError(t, err)
	assert.Equal(t, "Transfer", entry.Description)
	assert.Len(t, entry.Postings, 2)
	assert.Equal(t, decimal.NewFromFloat(100.00).String(), entry.Postings[0].Amount.String())
}
