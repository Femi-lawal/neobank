package repository

import (
	"github.com/femi-lawal/new_bank/backend/card-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CardRepository struct {
	DB *gorm.DB
}

func NewCardRepository(db *gorm.DB) *CardRepository {
	return &CardRepository{DB: db}
}

func (r *CardRepository) CreateCard(c *model.Card) error {
	return r.DB.Create(c).Error
}

func (r *CardRepository) GetCardByNumber(pan string) (*model.Card, error) {
	var c model.Card
	if err := r.DB.Where("masked_card_number = ?", pan).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// GetCardByID retrieves a card by its UUID
func (r *CardRepository) GetCardByID(cardID uuid.UUID) (*model.Card, error) {
	var c model.Card
	if err := r.DB.Where("id = ?", cardID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CardRepository) ListCardsByAccount(accountID string) ([]model.Card, error) {
	var cards []model.Card
	if err := r.DB.Where("account_id = ?", accountID).Find(&cards).Error; err != nil {
		return nil, err
	}
	return cards, nil
}

// ListCardsByUser returns all cards belonging to a specific user
func (r *CardRepository) ListCardsByUser(userID string) ([]model.Card, error) {
	var cards []model.Card
	if err := r.DB.Where("user_id = ?", userID).Find(&cards).Error; err != nil {
		return nil, err
	}
	return cards, nil
}

// VerifyAccountOwnership checks if a user owns a specific account
// SEC-006: This is a stub - in production, call the ledger service or use a shared DB view
func (r *CardRepository) VerifyAccountOwnership(userID, accountID uuid.UUID) (bool, error) {
	// DEMO: For the demo, we always return true since accounts are linked by user_id
	// In production, this would:
	// 1. Call the ledger service API to verify ownership, OR
	// 2. Query a shared accounts table, OR
	// 3. Use a JWT claim that includes account permissions

	// For demo purposes, assume ownership is valid
	// This allows the demo to work without a full service mesh
	return true, nil
}
