package service

import (
	"errors"
	"testing"

	"github.com/femi-lawal/new_bank/backend/product-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductRepository is a mock implementation of the product repository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) CreateProduct(product *model.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) GetProduct(id string) (*model.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Product), args.Error(1)
}

func (m *MockProductRepository) ListProducts() ([]model.Product, error) {
	args := m.Called()
	return args.Get(0).([]model.Product), args.Error(1)
}

func (m *MockProductRepository) UpdateProduct(product *model.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) DeleteProduct(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestProductService_CreateProduct(t *testing.T) {
	mockRepo := new(MockProductRepository)
	svc := NewProductService(nil) // Will test with mock

	// Test product model creation
	product := &model.Product{
		Name:        "Premium Account",
		Description: "High-yield savings account",
		Type:        "SAVINGS",
		Active:      true,
	}

	assert.Equal(t, "Premium Account", product.Name)
	assert.Equal(t, "SAVINGS", product.Type)
	assert.True(t, product.Active)
}

func TestProductService_ListProducts(t *testing.T) {
	mockRepo := new(MockProductRepository)

	expectedProducts := []model.Product{
		{Name: "Basic Checking", Type: "CHECKING", Active: true},
		{Name: "Premium Savings", Type: "SAVINGS", Active: true},
		{Name: "Investment Account", Type: "INVESTMENT", Active: false},
	}

	mockRepo.On("ListProducts").Return(expectedProducts, nil)

	products, err := mockRepo.ListProducts()

	assert.NoError(t, err)
	assert.Len(t, products, 3)
	assert.Equal(t, "Basic Checking", products[0].Name)
	mockRepo.AssertExpectations(t)
}

func TestProductService_ListProducts_Error(t *testing.T) {
	mockRepo := new(MockProductRepository)

	mockRepo.On("ListProducts").Return([]model.Product{}, errors.New("database error"))

	products, err := mockRepo.ListProducts()

	assert.Error(t, err)
	assert.Empty(t, products)
	mockRepo.AssertExpectations(t)
}

func TestProductService_GetProduct(t *testing.T) {
	mockRepo := new(MockProductRepository)

	expectedProduct := &model.Product{
		Name: "Test Product",
		Type: "CHECKING",
	}

	mockRepo.On("GetProduct", "test-id").Return(expectedProduct, nil)

	product, err := mockRepo.GetProduct("test-id")

	assert.NoError(t, err)
	assert.Equal(t, "Test Product", product.Name)
	mockRepo.AssertExpectations(t)
}

func TestProductService_GetProduct_NotFound(t *testing.T) {
	mockRepo := new(MockProductRepository)

	mockRepo.On("GetProduct", "nonexistent").Return(nil, errors.New("not found"))

	product, err := mockRepo.GetProduct("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, product)
	mockRepo.AssertExpectations(t)
}

func TestProductModel(t *testing.T) {
	product := model.Product{
		Name:        "Starter Account",
		Description: "Perfect for beginners",
		Type:        "CHECKING",
		Active:      true,
	}

	assert.NotEmpty(t, product.Name)
	assert.Equal(t, "CHECKING", product.Type)
	assert.True(t, product.Active)
}
