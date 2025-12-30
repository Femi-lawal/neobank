package service

import (
	"testing"

	"github.com/femi-lawal/new_bank/backend/payment-service/internal/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPaymentRepository is a mock implementation of the payment repository
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) CreatePayment(payment *model.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetPayment(id string) (*model.Payment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Payment), args.Error(1)
}

func (m *MockPaymentRepository) UpdatePaymentStatus(id string, status model.PaymentStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockPaymentRepository) ListPayments() ([]model.Payment, error) {
	args := m.Called()
	return args.Get(0).([]model.Payment), args.Error(1)
}

func TestInitiateTransfer_InvalidAmount(t *testing.T) {
	tests := []struct {
		name      string
		amount    string
		expectErr string
	}{
		{"empty amount", "", "invalid amount"},
		{"non-numeric", "abc", "invalid amount"},
		{"negative", "-100", "amount must be greater than zero"},
		{"zero", "0", "amount must be greater than zero"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &PaymentService{Repo: nil, useKafka: false}

			fromAcc := uuid.New().String()
			toAcc := uuid.New().String()

			_, err := svc.InitiateTransfer(fromAcc, toAcc, tt.amount, "USD", "test")

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectErr)
		})
	}
}

func TestInitiateTransfer_SameAccount(t *testing.T) {
	svc := &PaymentService{Repo: nil, useKafka: false}

	accountID := uuid.New().String()

	_, err := svc.InitiateTransfer(accountID, accountID, "100.00", "USD", "test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot transfer to the same account")
}

func TestInitiateTransfer_InvalidAccountIDs(t *testing.T) {
	svc := &PaymentService{Repo: nil, useKafka: false}

	validUUID := uuid.New().String()

	tests := []struct {
		name      string
		fromAcc   string
		toAcc     string
		expectErr string
	}{
		{"invalid from account", "not-a-uuid", validUUID, "invalid from account id"},
		{"invalid to account", validUUID, "not-a-uuid", "invalid to account id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.InitiateTransfer(tt.fromAcc, tt.toAcc, "100.00", "USD", "test")

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectErr)
		})
	}
}

func TestPaymentModel(t *testing.T) {
	payment := model.Payment{
		FromAccountID: uuid.New(),
		ToAccountID:   uuid.New(),
		Amount:        decimal.NewFromFloat(100.50),
		Currency:      "USD",
		Status:        model.StatusPending,
		Description:   "Test payment",
	}

	assert.Equal(t, model.StatusPending, payment.Status)
	assert.Equal(t, "USD", payment.Currency)
	assert.True(t, payment.Amount.Equal(decimal.NewFromFloat(100.50)))
}

func TestPaymentStatus(t *testing.T) {
	assert.Equal(t, model.PaymentStatus("PENDING"), model.StatusPending)
	assert.Equal(t, model.PaymentStatus("COMPLETED"), model.StatusCompleted)
	assert.Equal(t, model.PaymentStatus("FAILED"), model.StatusFailed)
}

func TestGetEnvOrDefault(t *testing.T) {
	// Test with unset variable
	result := getEnvOrDefault("NONEXISTENT_VAR_12345", "default")
	assert.Equal(t, "default", result)
}
