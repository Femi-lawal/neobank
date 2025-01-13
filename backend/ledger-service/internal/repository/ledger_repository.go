package repository

import (
	"errors"
	"fmt"

	"github.com/femi-lawal/new_bank/backend/ledger-service/internal/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type LedgerRepository struct {
	DB *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) *LedgerRepository {
	return &LedgerRepository{DB: db}
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
func (r *LedgerRepository) PostTransaction(entry *model.JournalEntry) error {
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

		// 3. Update Account Balances
		for _, p := range entry.Postings {
			// Lock account for update
			var account model.Account
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&account, "id = ?", p.AccountID).Error; err != nil {
				return fmt.Errorf("failed to lock account %s: %w", p.AccountID, err)
			}

			// Update balance
			movement := p.Amount
			if p.Direction == -1 {
				movement = movement.Neg()
			}

			// Simple signed balance update
			account.CachedBalance = account.CachedBalance.Add(movement)

			// Update Version
			account.BalanceVersion++

			if err := tx.Save(&account).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
