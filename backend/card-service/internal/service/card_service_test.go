package service

import (
	"errors"
	"os"
	"testing"

	"github.com/femi-lawal/new_bank/backend/card-service/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	// Set required env var for testing (32 bytes)
	os.Setenv("CARD_ENCRYPTION_KEY", "test_encryption_key_32_bytes_lni")
}

// MockCardRepository is a mock implementation of the card repository
type MockCardRepository struct {
	mock.Mock
}

func (m *MockCardRepository) CreateCard(card *model.Card) error {
	args := m.Called(card)
	return args.Error(0)
}

func (m *MockCardRepository) GetCardByNumber(pan string) (*model.Card, error) {
	args := m.Called(pan)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Card), args.Error(1)
}

func (m *MockCardRepository) GetCardByID(id uuid.UUID) (*model.Card, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Card), args.Error(1)
}

func (m *MockCardRepository) VerifyAccountOwnership(userID, accountID uuid.UUID) (bool, error) {
	args := m.Called(userID, accountID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCardRepository) ListCardsByAccount(accountID string) ([]model.Card, error) {
	args := m.Called(accountID)
	return args.Get(0).([]model.Card), args.Error(1)
}

func (m *MockCardRepository) ListCardsByUser(userID string) ([]model.Card, error) {
	args := m.Called(userID)
	return args.Get(0).([]model.Card), args.Error(1)
}

func TestCardService_IssueCard_InvalidUserID(t *testing.T) {
	svc := NewCardService(nil)

	_, err := svc.IssueCard("invalid-uuid", uuid.New().String())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user id")
}

func TestCardService_IssueCard_InvalidAccountID(t *testing.T) {
	svc := NewCardService(nil)

	_, err := svc.IssueCard(uuid.New().String(), "invalid-uuid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid account id")
}

func TestCardService_ListCardsByUser(t *testing.T) {
	mockRepo := new(MockCardRepository)
	svc := NewCardService(mockRepo) // Inject mock repo

	userID := uuid.New().String()
	expectedCards := []model.Card{
		{MaskedCardNumber: "**** **** **** 1111", Status: model.CardActive},
		{MaskedCardNumber: "**** **** **** 2222", Status: model.CardActive},
	}

	mockRepo.On("ListCardsByUser", userID).Return(expectedCards, nil)

	cards, err := svc.ListCardsByUser(userID) // Use service method

	assert.NoError(t, err)
	assert.Len(t, cards, 2)
	assert.Equal(t, "**** **** **** 1111", cards[0].MaskedCardNumber)
	mockRepo.AssertExpectations(t)
}

func TestCardService_ListCardsByUser_Empty(t *testing.T) {
	mockRepo := new(MockCardRepository)
	svc := NewCardService(mockRepo)

	userID := uuid.New().String()
	mockRepo.On("ListCardsByUser", userID).Return([]model.Card{}, nil)

	cards, err := svc.ListCardsByUser(userID)

	assert.NoError(t, err)
	assert.Empty(t, cards)
	mockRepo.AssertExpectations(t)
}

func TestCardService_ListCardsByUser_Error(t *testing.T) {
	mockRepo := new(MockCardRepository)
	svc := NewCardService(mockRepo)

	userID := uuid.New().String()
	mockRepo.On("ListCardsByUser", userID).Return([]model.Card{}, errors.New("database error"))

	cards, err := svc.ListCardsByUser(userID)

	assert.Error(t, err)
	assert.Empty(t, cards)
	mockRepo.AssertExpectations(t)
}

func TestCardModel(t *testing.T) {
	card := model.Card{
		UserID:              uuid.New(),
		AccountID:           uuid.New(),
		MaskedCardNumber:    "**** **** **** 1111",
		EncryptedCardNumber: "enc_data",
		ExpirationDate:      "12/27",
		Status:              model.CardActive,
	}

	assert.Equal(t, model.CardActive, card.Status)
	assert.Equal(t, "**** **** **** 1111", card.MaskedCardNumber)
	// Removed checks for CardNumber and CVV as they don't exist
}

func TestCardStatus(t *testing.T) {
	assert.Equal(t, model.CardStatus("ACTIVE"), model.CardActive)
	assert.Equal(t, model.CardStatus("BLOCKED"), model.CardBlocked)
	assert.Equal(t, model.CardStatus("INACTIVE"), model.CardInactive)
}

func TestGenerateRandomNumericString(t *testing.T) {
	// Test generating a 16-digit PAN
	pan, err := generateRandomNumericString(16)
	assert.NoError(t, err)
	assert.Len(t, pan, 16)

	// Verify all characters are digits
	for _, c := range pan {
		assert.True(t, c >= '0' && c <= '9')
	}

	// Test generating a 3-digit CVV
	cvv, err := generateRandomNumericString(3)
	assert.NoError(t, err)
	assert.Len(t, cvv, 3)
}
