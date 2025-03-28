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
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	AccountID uuid.UUID `gorm:"type:uuid;not null;index" json:"account_id"`
	// EncryptedCardNumber stores AES-256-GCM encrypted card number - NEVER exposed in API
	EncryptedCardNumber string `gorm:"column:encrypted_card_number;type:text;not null" json:"-"`
	// MaskedCardNumber stores only displayable format: **** **** **** 1234
	MaskedCardNumber string `gorm:"column:masked_card_number;type:varchar(19);not null" json:"card_number"`
	// CVV is NEVER stored per PCI DSS 3.2 - only used for single-transaction validation
	ExpirationDate string     `gorm:"type:varchar(5);not null" json:"expiration_date"` // MM/YY
	Status         CardStatus `gorm:"type:varchar(20);default:'ACTIVE'" json:"status"`
	// CardToken for payment processing - replaces actual card number in transactions
	CardToken  uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid()" json:"card_token"`
	PinHash    string          `gorm:"type:varchar(255)" json:"-"` // Never expose PIN
	DailyLimit decimal.Decimal `gorm:"type:numeric(19,4);default:1000.00" json:"daily_limit"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	DeletedAt  gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName specifies the table name for GORM
func (Card) TableName() string {
	return "cards"
}
