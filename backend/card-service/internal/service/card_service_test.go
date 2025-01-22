package service

import (
	"errors"
	"testing"

	"github.com/femi-lawal/new_bank/backend/card-service/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func (m *MockCardRepository) ListCardsByAccount(accountID string) ([]model.Card, error) {
	args := m.Called(accountID)
	return args.Get(0).([]model.Card), args.Error(1)
}

func (m *MockCardRepository) ListCardsByUser(userID string) ([]model.Card, error) {
	args := m.Called(userID)
	return args.Get(0).([]model.Card), args.Error(1)
}

func TestCardService_IssueCard_InvalidUserID(t *testing.T) {
	mockRepo := new(MockCardRepository)
	svc := NewCardService(nil)

	_, err := svc.IssueCard("invalid-uuid", uuid.New().String())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user id")
}

func TestCardService_IssueCard_InvalidAccountID(t *testing.T) {
	mockRepo := new(MockCardRepository)
	svc := NewCardService(nil)

	_, err := svc.IssueCard(uuid.New().String(), "invalid-uuid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid account id")
}

func TestCardService_ListCardsByUser(t *testing.T) {
	mockRepo := new(MockCardRepository)

	userID := uuid.New().String()
	expectedCards := []model.Card{
		{CardNumber: "4111111111111111", Status: model.CardActive},
		{CardNumber: "4222222222222222", Status: model.CardActive},
	}

	mockRepo.On("ListCardsByUser", userID).Return(expectedCards, nil)

	cards, err := mockRepo.ListCardsByUser(userID)

	assert.NoError(t, err)
	assert.Len(t, cards, 2)
	mockRepo.AssertExpectations(t)
}

func TestCardService_ListCardsByUser_Empty(t *testing.T) {
	mockRepo := new(MockCardRepository)

	userID := uuid.New().String()
	mockRepo.On("ListCardsByUser", userID).Return([]model.Card{}, nil)

	cards, err := mockRepo.ListCardsByUser(userID)

	assert.NoError(t, err)
	assert.Empty(t, cards)
	mockRepo.AssertExpectations(t)
}

func TestCardService_ListCardsByUser_Error(t *testing.T) {
	mockRepo := new(MockCardRepository)

	userID := uuid.New().String()
	mockRepo.On("ListCardsByUser", userID).Return([]model.Card{}, errors.New("database error"))

	cards, err := mockRepo.ListCardsByUser(userID)

	assert.Error(t, err)
	assert.Empty(t, cards)
	mockRepo.AssertExpectations(t)
}

func TestCardModel(t *testing.T) {
	card := model.Card{
		UserID:         uuid.New(),
		AccountID:      uuid.New(),
		CardNumber:     "4111111111111111",
		CVV:            "123",
		ExpirationDate: "12/27",
		Status:         model.CardActive,
	}

	assert.Equal(t, model.CardActive, card.Status)
	assert.Len(t, card.CardNumber, 16)
	assert.Len(t, card.CVV, 3)
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
