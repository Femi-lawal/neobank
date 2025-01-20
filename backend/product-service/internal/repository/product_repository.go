package repository

import (
	"github.com/femi-lawal/new_bank/backend/product-service/internal/model"
	"gorm.io/gorm"
)

type ProductRepository struct {
	DB *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{DB: db}
}

func (r *ProductRepository) CreateProduct(p *model.Product) error {
	return r.DB.Create(p).Error
}

func (r *ProductRepository) ListProducts() ([]model.Product, error) {
	var products []model.Product
	if err := r.DB.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepository) GetProductByCode(code string) (*model.Product, error) {
	var p model.Product
	if err := r.DB.Where("code = ?", code).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}
