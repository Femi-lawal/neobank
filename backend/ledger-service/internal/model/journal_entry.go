package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type JournalEntryStatus string

const (
	StatusPending JournalEntryStatus = "PENDING"
	StatusPosted  JournalEntryStatus = "POSTED"
	StatusVoid    JournalEntryStatus = "VOID"
)

type JournalEntry struct {
	ID              uuid.UUID          `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TransactionDate time.Time          `gorm:"not null"`
	Description     string             `gorm:"type:text"`
	ReferenceID     string             `gorm:"type:varchar(100);index"`
	Status          JournalEntryStatus `gorm:"type:varchar(20);default:'POSTED'"`
	Postings        []Posting          `gorm:"foreignKey:JournalEntryID"`
	CreatedAt       time.Time
}

type Posting struct {
	ID             uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	JournalEntryID uuid.UUID       `gorm:"type:uuid;not null;index"`
	AccountID      uuid.UUID       `gorm:"type:uuid;not null;index"`
	Amount         decimal.Decimal `gorm:"type:numeric(19,4);not null;check:amount > 0"`
	Direction      int             `gorm:"type:smallint;not null;check:direction IN (1, -1)"` // 1 = Debit, -1 = Credit
	CreatedAt      time.Time
}
