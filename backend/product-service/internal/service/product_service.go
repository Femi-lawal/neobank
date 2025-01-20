package service

import (
	"github.com/femi-lawal/new_bank/backend/product-service/internal/model"
	"github.com/femi-lawal/new_bank/backend/product-service/internal/repository"
	"github.com/shopspring/decimal"
)

type ProductService struct {
	Repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{Repo: repo}
}

func (s *ProductService) CreateProduct(code, name string, pType model.ProductType, interestRateStr string, currency string) (*model.Product, error) {
	rate, err := decimal.NewFromString(interestRateStr)
	if err != nil {
		return nil, err
	}

	p := &model.Product{
		Code:         code,
		Name:         name,
		Type:         pType,
		InterestRate: rate,
		CurrencyCode: currency,
	}

	if err := s.Repo.CreateProduct(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProductService) ListProducts() ([]model.Product, error) {
	return s.Repo.ListProducts()
}
