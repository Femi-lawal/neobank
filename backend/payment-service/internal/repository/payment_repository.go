package repository

import (
	"github.com/femi-lawal/new_bank/backend/payment-service/internal/model"
	"gorm.io/gorm"
)

type PaymentRepository struct {
	DB *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{DB: db}
}

func (r *PaymentRepository) CreatePayment(p *model.Payment) error {
	return r.DB.Create(p).Error
}

func (r *PaymentRepository) UpdateStatus(id string, status model.PaymentStatus) error {
	return r.DB.Model(&model.Payment{}).Where("id = ?", id).Update("status", status).Error
}

func (r *PaymentRepository) GetPayment(id string) (*model.Payment, error) {
	var p model.Payment
	if err := r.DB.Where("id = ?", id).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}
