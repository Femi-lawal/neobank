package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PaymentStatus string

const (
	StatusPending   PaymentStatus = "PENDING"
	StatusCompleted PaymentStatus = "COMPLETED"
	StatusFailed    PaymentStatus = "FAILED"
)

type Payment struct {
	ID            uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FromAccountID uuid.UUID       `gorm:"type:uuid;not null"`
	ToAccountID   uuid.UUID       `gorm:"type:uuid;not null"`
	Amount        decimal.Decimal `gorm:"type:numeric(19,4);not null"`
	Currency      string          `gorm:"type:char(3);not null"`
	Status        PaymentStatus   `gorm:"type:varchar(20);default:'PENDING'"`
	Description   string          `gorm:"type:text"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}
