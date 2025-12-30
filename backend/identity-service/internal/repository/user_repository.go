package repository

import (
	"github.com/femi-lawal/new_bank/backend/identity-service/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID finds a user by their ID
func (r *UserRepository) FindByID(id string) (*model.User, error) {
	var user model.User
	if err := r.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdatePassword updates a user's password hash
func (r *UserRepository) UpdatePassword(userID string, hashedPassword string) error {
	return r.DB.Model(&model.User{}).Where("id = ?", userID).Update("password_hash", hashedPassword).Error
}
