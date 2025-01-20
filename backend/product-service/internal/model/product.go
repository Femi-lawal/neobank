package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ProductType string

const (
	Savings  ProductType = "SAVINGS"
	Checking ProductType = "CHECKING"
	Loan     ProductType = "LOAN"
)

type Product struct {
	ID           uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Code         string          `gorm:"uniqueIndex;not null;type:varchar(50)"` // e.g., "SAVINGS-STD"
	Name         string          `gorm:"type:varchar(100);not null"`
	Type         ProductType     `gorm:"type:varchar(20);not null"`
	InterestRate decimal.Decimal `gorm:"type:numeric(5,4);default:0"` // e.g. 0.0500 for 5%
	CurrencyCode string          `gorm:"type:char(3);not null"`
	Metadata     *string         `gorm:"type:jsonb"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
