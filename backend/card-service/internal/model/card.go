package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CardStatus string

const (
	CardActive   CardStatus = "ACTIVE"
	CardBlocked  CardStatus = "BLOCKED"
	CardInactive CardStatus = "INACTIVE"
)

type Card struct {
	ID             uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID       `gorm:"type:uuid;not null;index" json:"user_id"`
	AccountID      uuid.UUID       `gorm:"type:uuid;not null;index" json:"account_id"`
	CardNumber     string          `gorm:"type:varchar(16);uniqueIndex;not null" json:"card_number"` // Last 4 visible in real app
	CVV            string          `gorm:"type:varchar(4);not null" json:"-"`                        // Never expose CVV
	ExpirationDate string          `gorm:"type:varchar(5);not null" json:"expiration_date"`          // MM/YY
	Status         CardStatus      `gorm:"type:varchar(20);default:'ACTIVE'" json:"status"`
	PinHash        string          `gorm:"type:varchar(255)" json:"-"` // Never expose PIN
	DailyLimit     decimal.Decimal `gorm:"type:numeric(19,4);default:1000.00" json:"daily_limit"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
}
