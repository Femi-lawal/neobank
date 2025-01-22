package repository

import (
	"github.com/femi-lawal/new_bank/backend/card-service/internal/model"
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
	if err := r.DB.Where("card_number = ?", pan).First(&c).Error; err != nil {
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
