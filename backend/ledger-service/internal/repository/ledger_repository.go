package repository

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/femi-lawal/new_bank/backend/ledger-service/internal/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// MaxRetries is the maximum number of retries for serialization/deadlock errors
const MaxRetries = 3

// RetryableErrors are PostgreSQL error codes that should trigger a retry
var RetryableErrors = []string{
	"40001", // serialization_failure
	"40P01", // deadlock_detected
}

type LedgerRepository struct {
	DB *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) *LedgerRepository {
	return &LedgerRepository{DB: db}
}

// isRetryableError checks if the error is a retryable PostgreSQL error
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	for _, code := range RetryableErrors {
		if strings.Contains(errStr, code) {
			return true
		}
	}
	return false
}

func (r *LedgerRepository) CreateAccount(account *model.Account) error {
	return r.DB.Create(account).Error
}

func (r *LedgerRepository) GetAccount(id string) (*model.Account, error) {
	var account model.Account
	if err := r.DB.Where("id = ?", id).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *LedgerRepository) ListAccounts() ([]model.Account, error) {
	var accounts []model.Account
	if err := r.DB.Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

// ListAccountsByUser returns accounts for a specific user
func (r *LedgerRepository) ListAccountsByUser(userID string) ([]model.Account, error) {
	var accounts []model.Account
	if err := r.DB.Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

// PostTransaction executes a journal entry and updates balances atomically using Database Transaction.
// Implements retry logic for serialization failures and deadlocks, with deterministic lock ordering.
func (r *LedgerRepository) PostTransaction(entry *model.JournalEntry) error {
	var lastErr error
	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter
			backoffMs := (1 << attempt) * 50 // 100ms, 200ms, 400ms
			time.Sleep(time.Duration(backoffMs) * time.Millisecond)
			slog.Info("Retrying transaction", "attempt", attempt+1, "lastError", lastErr)
		}

		lastErr = r.postTransactionOnce(entry)
		if lastErr == nil {
			return nil
		}

		if !isRetryableError(lastErr) {
			return lastErr // Non-retryable error, return immediately
		}
	}
	return fmt.Errorf("transaction failed after %d retries: %w", MaxRetries, lastErr)
}

// postTransactionOnce executes the transaction once (called by PostTransaction with retry logic)
func (r *LedgerRepository) postTransactionOnce(entry *model.JournalEntry) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Validate Double Entry (Sum of Debits == Sum of Credits)
		// Actually, in signed ledger: Sum(Amount * Direction) == 0
		var sum decimal.Decimal
		for _, p := range entry.Postings {
			amount := p.Amount
			if p.Direction == -1 {
				amount = amount.Neg()
			}
			sum = sum.Add(amount)
		}

		if !sum.IsZero() {
			return errors.New("transaction is not balanced")
		}

		// 2. Create Journal Entry
		if err := tx.Create(entry).Error; err != nil {
			return err
		}

		// 3. Collect and sort account IDs for deterministic lock ordering (prevents deadlocks)
		accountIDs := make([]string, len(entry.Postings))
		for i, p := range entry.Postings {
			accountIDs[i] = p.AccountID.String()
		}
		sort.Strings(accountIDs)

		// Create a map of account ID to posting for efficient lookup
		postingMap := make(map[string][]model.Posting)
		for _, p := range entry.Postings {
			id := p.AccountID.String()
			postingMap[id] = append(postingMap[id], p)
		}

		// 4. Lock and update accounts in sorted order to prevent deadlocks
		for _, accID := range accountIDs {
			// Lock account for update (deterministic order)
			var account model.Account
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&account, "id = ?", accID).Error; err != nil {
				return fmt.Errorf("failed to lock account %s: %w", accID, err)
			}

			// Apply all postings for this account
			for _, p := range postingMap[accID] {
				// Update balance
				movement := p.Amount
				if p.Direction == -1 {
					movement = movement.Neg()
				}
				account.CachedBalance = account.CachedBalance.Add(movement)
			}

			// Update Version
			account.BalanceVersion++

			if err := tx.Save(&account).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
