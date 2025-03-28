package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/femi-lawal/new_bank/backend/card-service/internal/model"
	"github.com/femi-lawal/new_bank/backend/card-service/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrUnauthorized     = errors.New("unauthorized: you do not own this account")
	ErrInvalidUserID    = errors.New("invalid user id")
	ErrInvalidAccountID = errors.New("invalid account id")
)

type CardService struct {
	Repo *repository.CardRepository
}

func NewCardService(repo *repository.CardRepository) *CardService {
	return &CardService{Repo: repo}
}

// IssueCard creates a new card for the authenticated user
// SEC-006: Validates that the user owns the account before issuing a card
func (s *CardService) IssueCard(userID, accountID string) (*model.Card, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	accUUID, err := uuid.Parse(accountID)
	if err != nil {
		return nil, ErrInvalidAccountID
	}

	// SEC-006: Verify the user owns the account before proceeding
	// In production, this would call the ledger service to verify ownership
	ownsAccount, err := s.Repo.VerifyAccountOwnership(userUUID, accUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify account ownership: %w", err)
	}
	if !ownsAccount {
		return nil, ErrUnauthorized
	}

	// Generate Random PAN (Mock - in production use payment processor)
	pan, _ := generateRandomNumericString(16)

	// SEC-002: CVV is NEVER stored - only generated for single-use display
	// In real implementation, CVV would be shown once and never stored

	// Expiry +3 years
	expiry := time.Now().AddDate(3, 0, 0).Format("01/06")

	// SEC-003: Encrypt card number for storage
	encryptedPAN := encryptCardNumber(pan)
	maskedPAN := maskCardNumber(pan)

	card := &model.Card{
		UserID:              userUUID,
		AccountID:           accUUID,
		EncryptedCardNumber: encryptedPAN,
		MaskedCardNumber:    maskedPAN,
		ExpirationDate:      expiry,
		Status:              model.CardActive,
		CardToken:           uuid.New(),
		DailyLimit:          decimal.NewFromInt(1000),
	}

	if err := s.Repo.CreateCard(card); err != nil {
		return nil, err
	}
	return card, nil
}

// ListCards returns cards for a specific account
// SEC-006: Should be called after verifying user owns the account
func (s *CardService) ListCards(accountID string) ([]model.Card, error) {
	return s.Repo.ListCardsByAccount(accountID)
}

// ListCardsByUser returns all cards belonging to a user
func (s *CardService) ListCardsByUser(userID string) ([]model.Card, error) {
	return s.Repo.ListCardsByUser(userID)
}

// GetCard retrieves a specific card with ownership validation
// SEC-006: Validates that the requesting user owns the card
func (s *CardService) GetCard(userID, cardID string) (*model.Card, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	cardUUID, err := uuid.Parse(cardID)
	if err != nil {
		return nil, errors.New("invalid card id")
	}

	card, err := s.Repo.GetCardByID(cardUUID)
	if err != nil {
		return nil, err
	}

	// SEC-006: Verify the user owns this card
	if card.UserID != userUUID {
		return nil, ErrUnauthorized
	}

	return card, nil
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

// maskCardNumber returns a masked version of the card number
// Format: **** **** **** 1234
func maskCardNumber(pan string) string {
	if len(pan) < 4 {
		return "****"
	}
	return fmt.Sprintf("**** **** **** %s", pan[len(pan)-4:])
}

// encryptCardNumber encrypts the card number for storage
// SEC-003: In production, use AES-256-GCM with KMS-managed keys
func encryptCardNumber(pan string) string {
	// DEMO: Base64 placeholder - in production use real encryption
	// Example: return aes256gcm.Encrypt(pan, kmsKey)
	return fmt.Sprintf("ENC[%s]", pan) // Placeholder for demo
}
