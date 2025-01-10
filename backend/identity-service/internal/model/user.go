package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email        string         `gorm:"uniqueIndex;not null"`
	PasswordHash string         `gorm:"not null"`
	FirstName    string         `gorm:"not null"`
	LastName     string         `gorm:"not null"`
	Role         string         `gorm:"default:'customer'"`
	KYCStatus    string         `gorm:"default:'UNVERIFIED'"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
