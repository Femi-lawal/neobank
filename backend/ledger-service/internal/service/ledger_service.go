package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/femi-lawal/new_bank/backend/ledger-service/internal/model"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/cache"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LedgerRepository interface {
	CreateAccount(acc *model.Account) error
	GetAccount(id string) (*model.Account, error)
	ListAccounts() ([]model.Account, error)
	ListAccountsByUser(userID string) ([]model.Account, error)
	PostTransaction(entry *model.JournalEntry) error
}

type LedgerService struct {
	Repo  LedgerRepository
	cache *cache.RedisClient
}

// NewLedgerService creates a ledger service without caching
func NewLedgerService(repo LedgerRepository) *LedgerService {
	return &LedgerService{Repo: repo}
}

// NewLedgerServiceWithCache creates a ledger service with Redis caching
func NewLedgerServiceWithCache(repo LedgerRepository, redisClient *cache.RedisClient) *LedgerService {
	return &LedgerService{Repo: repo, cache: redisClient}
}

func (s *LedgerService) CreateAccount(userID, accountNumber, name, currency string, accType model.AccountType) (*model.Account, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	acc := &model.Account{
		UserID:        userUUID,
		AccountNumber: accountNumber,
		Name:          name,
		Type:          accType,
		CurrencyCode:  currency,
		CachedBalance: decimal.Zero,
	}
	if err := s.Repo.CreateAccount(acc); err != nil {
		return nil, err
	}

	// Invalidate account list cache
	if s.cache != nil {
		s.cache.Delete(context.Background(), "accounts:list:"+userID)
	}

	return acc, nil
}

// ListAccountsByUser returns accounts for a specific user
func (s *LedgerService) ListAccountsByUser(userID string) ([]model.Account, error) {
	cacheKey := "accounts:list:" + userID

	// Try cache first
	if s.cache != nil {
		var accounts []model.Account
		err := s.cache.GetJSON(context.Background(), cacheKey, &accounts)
		if err == nil && len(accounts) > 0 {
			slog.Debug("Cache hit for user accounts list", "user_id", userID)
			return accounts, nil
		}
	}

	// Fallback to DB
	accounts, err := s.Repo.ListAccountsByUser(userID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if s.cache != nil && len(accounts) > 0 {
		s.cache.SetJSON(context.Background(), cacheKey, accounts, cache.DefaultCacheTTL)
	}

	return accounts, nil
}

func (s *LedgerService) ListAccounts() ([]model.Account, error) {
	// Try cache first
	if s.cache != nil {
		var accounts []model.Account
		err := s.cache.GetJSON(context.Background(), "accounts:list", &accounts)
		if err == nil && len(accounts) > 0 {
			slog.Debug("Cache hit for accounts list")
			return accounts, nil
		}
	}

	// Fallback to DB
	accounts, err := s.Repo.ListAccounts()
	if err != nil {
		return nil, err
	}

	// Cache the result
	if s.cache != nil && len(accounts) > 0 {
		s.cache.SetJSON(context.Background(), "accounts:list", accounts, cache.DefaultCacheTTL)
	}

	return accounts, nil
}

type PostingRequest struct {
	AccountID string
	Amount    string
	Direction int
}

// PostTransaction creates a journal entry with multiple postings
func (s *LedgerService) PostTransaction(desc string, postings []PostingRequest) (*model.JournalEntry, error) {
	if len(postings) < 2 {
		return nil, errors.New("transaction must have at least 2 postings")
	}

	entry := &model.JournalEntry{
		TransactionDate: time.Now(),
		Description:     desc,
		Status:          model.StatusPosted,
		Postings:        make([]model.Posting, len(postings)),
	}

	var affectedAccounts []string

	for i, p := range postings {
		amount, err := decimal.NewFromString(p.Amount)
		if err != nil {
			return nil, errors.New("invalid amount format")
		}

		accUUID, err := uuid.Parse(p.AccountID)
		if err != nil {
			return nil, errors.New("invalid account UUID")
		}

		entry.Postings[i] = model.Posting{
			AccountID: accUUID,
			Amount:    amount,
			Direction: p.Direction,
		}
		affectedAccounts = append(affectedAccounts, p.AccountID)
	}

	if err := s.Repo.PostTransaction(entry); err != nil {
		return nil, err
	}

	// Invalidate cache for affected accounts
	if s.cache != nil {
		ctx := context.Background()
		for _, accID := range affectedAccounts {
			s.cache.Delete(ctx, cache.BalanceCacheKey(accID))
			s.cache.Delete(ctx, cache.AccountCacheKey(accID))

			// Also invalidate cache for the user who owns this account
			acc, err := s.Repo.GetAccount(accID)
			if err == nil && acc != nil {
				slog.Debug("Invalidating cache for user", "user_id", acc.UserID.String())
				s.cache.Delete(ctx, "accounts:list:"+acc.UserID.String())
			}
		}
		// Also invalidate accounts list since balances changed
		s.cache.Delete(ctx, "accounts:list")
		slog.Debug("Cache invalidated for accounts", "count", len(affectedAccounts))
	}

	return entry, nil
}

// PostTransfer is a convenience method for simple A->B transfers (used by Kafka consumer)
func (s *LedgerService) PostTransfer(fromAccountID, toAccountID, amountStr, description string) (*model.JournalEntry, error) {
	postings := []PostingRequest{
		{AccountID: fromAccountID, Amount: amountStr, Direction: -1}, // Credit sender
		{AccountID: toAccountID, Amount: amountStr, Direction: 1},    // Debit receiver
	}
	return s.PostTransaction(description, postings)
}
