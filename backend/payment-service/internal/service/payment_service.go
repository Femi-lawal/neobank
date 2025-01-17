package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/femi-lawal/new_bank/backend/payment-service/internal/model"
	"github.com/femi-lawal/new_bank/backend/payment-service/internal/repository"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/kafka"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentService struct {
	Repo      *repository.PaymentRepository
	producer  *kafka.Producer
	useKafka  bool
	ledgerURL string // Configurable ledger service URL
}

// NewPaymentService creates a new payment service (sync mode - fallback)
func NewPaymentService(repo *repository.PaymentRepository) *PaymentService {
	return &PaymentService{
		Repo:      repo,
		useKafka:  false,
		ledgerURL: getEnvOrDefault("LEDGER_SERVICE_URL", "http://localhost:8082"),
	}
}

// NewPaymentServiceWithKafka creates a payment service with Kafka async processing
func NewPaymentServiceWithKafka(repo *repository.PaymentRepository, producer *kafka.Producer) *PaymentService {
	return &PaymentService{
		Repo:      repo,
		producer:  producer,
		useKafka:  true,
		ledgerURL: getEnvOrDefault("LEDGER_SERVICE_URL", "http://localhost:8082"),
	}
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type LedgerTransactionRequest struct {
	Description string `json:"description"`
	Postings    []struct {
		AccountID string `json:"account_id"`
		Amount    string `json:"amount"`
		Direction int    `json:"direction"`
	} `json:"postings"`
}

func (s *PaymentService) InitiateTransfer(fromAcc, toAcc, amountStr, currency, desc string) (*model.Payment, error) {
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return nil, errors.New("invalid amount")
	}

	// Validate amount is positive
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("amount must be greater than zero")
	}

	// Check for same account transfer
	if fromAcc == toAcc {
		return nil, errors.New("cannot transfer to the same account")
	}

	fromUUID, err := uuid.Parse(fromAcc)
	if err != nil {
		return nil, errors.New("invalid from account id")
	}
	toUUID, err := uuid.Parse(toAcc)
	if err != nil {
		return nil, errors.New("invalid to account id")
	}

	// Validate balance by calling ledger service
	balanceErr := s.validateBalance(fromAcc, amountStr)
	if balanceErr != nil {
		return nil, balanceErr
	}

	// 1. Create Pending Payment
	payment := &model.Payment{
		FromAccountID: fromUUID,
		ToAccountID:   toUUID,
		Amount:        amount,
		Currency:      currency,
		Status:        model.StatusPending,
		Description:   desc,
	}

	if err := s.Repo.CreatePayment(payment); err != nil {
		return nil, err
	}

	// 2. Process transfer - async via Kafka or sync via HTTP
	if s.useKafka && s.producer != nil {
		// Async: Publish to Kafka and return immediately
		return s.processAsync(payment, fromAcc, toAcc, amountStr, currency, desc)
	}

	// Sync: Call Ledger Service directly (fallback)
	return s.processSync(payment, fromAcc, toAcc, amountStr, desc)
}

// processAsync publishes payment event to Kafka for async processing
func (s *PaymentService) processAsync(payment *model.Payment, fromAcc, toAcc, amountStr, currency, desc string) (*model.Payment, error) {
	event := kafka.PaymentEvent{
		PaymentID:     payment.ID.String(),
		FromAccountID: fromAcc,
		ToAccountID:   toAcc,
		Amount:        amountStr,
		Currency:      currency,
		Description:   desc,
		Status:        string(model.StatusPending),
		Timestamp:     time.Now().Format(time.RFC3339),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.producer.Produce(ctx, kafka.TopicPaymentCreated, payment.ID.String(), event)
	if err != nil {
		slog.Error("Failed to publish payment event to Kafka", "payment_id", payment.ID, "error", err)
		// Fallback to sync processing
		return s.processSync(payment, fromAcc, toAcc, amountStr, desc)
	}

	slog.Info("Payment event published to Kafka", "payment_id", payment.ID, "topic", kafka.TopicPaymentCreated)

	// Return immediately with PENDING status - ledger service will process async
	return payment, nil
}

// processSync calls ledger service synchronously (original behavior)
func (s *PaymentService) processSync(payment *model.Payment, fromAcc, toAcc, amountStr, desc string) (*model.Payment, error) {
	err := s.callLedger(fromAcc, toAcc, amountStr, desc)
	if err != nil {
		s.Repo.UpdateStatus(payment.ID.String(), model.StatusFailed)
		return payment, fmt.Errorf("ledger transfer failed: %w", err)
	}

	// Mark Complete
	s.Repo.UpdateStatus(payment.ID.String(), model.StatusCompleted)
	payment.Status = model.StatusCompleted

	return payment, nil
}

// UpdatePaymentStatus updates payment status (called by consumer after processing)
func (s *PaymentService) UpdatePaymentStatus(paymentID string, status model.PaymentStatus) error {
	return s.Repo.UpdateStatus(paymentID, status)
}

func (s *PaymentService) callLedger(from, to, amount, desc string) error {
	req := LedgerTransactionRequest{
		Description: "Payment: " + desc,
		Postings: []struct {
			AccountID string `json:"account_id"`
			Amount    string `json:"amount"`
			Direction int    `json:"direction"`
		}{
			{AccountID: from, Amount: amount, Direction: -1}, // Credit Sender
			{AccountID: to, Amount: amount, Direction: 1},    // Debit Receiver
		},
	}

	body, _ := json.Marshal(req)
	url := s.ledgerURL + "/api/v1/transactions"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return errors.New("ledger service returned non-201 status")
	}

	return nil
}

// AccountResponse represents the account data from ledger service
type AccountResponse struct {
	ID      string `json:"id"`
	Balance string `json:"balance"`
}

// validateBalance checks if the from account has sufficient balance for the transfer
func (s *PaymentService) validateBalance(fromAccountID, amountStr string) error {
	// Call ledger service to get account balance
	url := s.ledgerURL + "/api/v1/accounts/" + fromAccountID
	resp, err := http.Get(url)
	if err != nil {
		// If we can't verify balance, log warning but allow transfer (may fail at ledger level)
		slog.Warn("Could not verify balance, proceeding with transfer", "account", fromAccountID, "error", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If account not found, the transfer will fail anyway at ledger level
		slog.Warn("Account not found or ledger error", "account", fromAccountID, "status", resp.StatusCode)
		return nil
	}

	var account AccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		slog.Warn("Could not decode account response", "error", err)
		return nil
	}

	balance, err := decimal.NewFromString(account.Balance)
	if err != nil {
		slog.Warn("Could not parse account balance", "balance", account.Balance, "error", err)
		return nil
	}

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return errors.New("invalid amount")
	}

	if balance.LessThan(amount) {
		return fmt.Errorf("insufficient funds: available %s, requested %s", balance.String(), amount.String())
	}

	return nil
}
