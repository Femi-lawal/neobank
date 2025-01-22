package service

import (
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/femi-lawal/new_bank/backend/card-service/internal/model"
	"github.com/femi-lawal/new_bank/backend/card-service/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CardService struct {
	Repo *repository.CardRepository
}

func NewCardService(repo *repository.CardRepository) *CardService {
	return &CardService{Repo: repo}
}

// IssueCard creates a new card for the authenticated user
func (s *CardService) IssueCard(userID, accountID string) (*model.Card, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	accUUID, err := uuid.Parse(accountID)
	if err != nil {
		return nil, errors.New("invalid account id")
	}

	// Generate Random PAN (Mock)
	pan, _ := generateRandomNumericString(16)
	cvv, _ := generateRandomNumericString(3)

	// Expiry +3 years
	expiry := time.Now().AddDate(3, 0, 0).Format("01/06")

	card := &model.Card{
		UserID:         userUUID,
		AccountID:      accUUID,
		CardNumber:     pan,
		CVV:            cvv,
		ExpirationDate: expiry,
		Status:         model.CardActive,
		DailyLimit:     decimal.NewFromInt(1000),
	}

	if err := s.Repo.CreateCard(card); err != nil {
		return nil, err
	}
	return card, nil
}

// ListCards returns cards for a specific account (legacy method)
func (s *CardService) ListCards(accountID string) ([]model.Card, error) {
	return s.Repo.ListCardsByAccount(accountID)
}

// ListCardsByUser returns all cards belonging to a user
func (s *CardService) ListCardsByUser(userID string) ([]model.Card, error) {
	return s.Repo.ListCardsByUser(userID)
}

func generateRandomNumericString(n int) (string, error) {
	const letters = "0123456789"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret), nil
}
