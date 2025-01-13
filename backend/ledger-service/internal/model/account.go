package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AccountType string

const (
	Asset     AccountType = "ASSET"
	Liability AccountType = "LIABILITY"
	Equity    AccountType = "EQUITY"
	Income    AccountType = "INCOME"
	Expense   AccountType = "EXPENSE"
)

type Account struct {
	ID             uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID       `gorm:"type:uuid;index;not null" json:"user_id"`
	AccountNumber  string          `gorm:"uniqueIndex;not null;type:varchar(20)" json:"account_number"`
	Name           string          `gorm:"type:varchar(100)" json:"name"`
	Type           AccountType     `gorm:"type:varchar(20);not null" json:"type"`
	CurrencyCode   string          `gorm:"type:char(3);not null" json:"currency_code"`
	Status         string          `gorm:"type:varchar(20);default:'ACTIVE'" json:"status"`
	BalanceVersion int             `gorm:"default:0" json:"-"`
	CachedBalance  decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"balance"`
	Metadata       *string         `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
}
