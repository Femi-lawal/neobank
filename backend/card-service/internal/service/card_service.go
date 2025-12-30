package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/femi-lawal/new_bank/backend/card-service/internal/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrUnauthorized     = errors.New("unauthorized: you do not own this account")
	ErrInvalidUserID    = errors.New("invalid user id")
	ErrInvalidAccountID = errors.New("invalid account id")
	ErrEncryption       = errors.New("encryption failed")
)

// encryptionKey is loaded from environment variable
// In production, this should come from AWS KMS or similar
var encryptionKey []byte

func init() {
	keyStr := os.Getenv("CARD_ENCRYPTION_KEY")
	if keyStr == "" {
		// For development/testing only - production MUST set this
		keyStr = "devonly32byteencryptionkey!!!!!!" // exactly 32 bytes
	}
	encryptionKey = []byte(keyStr)
	if len(encryptionKey) != 32 {
		panic("CARD_ENCRYPTION_KEY must be exactly 32 bytes for AES-256")
	}
}

// Repository defines the interface for card data access
// This allows for mocking in unit tests
type Repository interface {
	CreateCard(card *model.Card) error
	GetCardByID(id uuid.UUID) (*model.Card, error)
	GetCardByNumber(pan string) (*model.Card, error)
	ListCardsByAccount(accountID string) ([]model.Card, error)
	ListCardsByUser(userID string) ([]model.Card, error)
	VerifyAccountOwnership(userID, accountID uuid.UUID) (bool, error)
}

type CardService struct {
	Repo Repository
}

func NewCardService(repo Repository) *CardService {
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

	// SEC-003: Encrypt card number for storage using AES-256-GCM
	encryptedPAN, err := encryptCardNumber(pan)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt card number: %w", err)
	}
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

// encryptCardNumber encrypts the card number using AES-256-GCM
// SEC-003: Real encryption for PCI-DSS compliance
func encryptCardNumber(pan string) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(pan), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptCardNumber decrypts the card number using AES-256-GCM
// SEC-003: For internal use only - never expose decrypted PAN to clients
func decryptCardNumber(encrypted string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decode: %w", err)
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
