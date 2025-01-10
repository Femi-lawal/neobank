package service

import (
	"errors"
	"time"

	"github.com/femi-lawal/new_bank/backend/identity-service/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	FindByEmail(email string) (*model.User, error)
	Create(user *model.User) error
}

type AuthService struct {
	Repo      UserRepository
	JWTSecret []byte
}

func NewAuthService(repo UserRepository, secret string) *AuthService {
	return &AuthService{
		Repo:      repo,
		JWTSecret: []byte(secret),
	}
}

func (s *AuthService) Register(email, password, firstName, lastName string) (*model.User, error) {
	// Check if user exists
	if _, err := s.Repo.FindByEmail(email); err == nil {
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    firstName,
		LastName:     lastName,
		Role:         "customer",
	}

	if err := s.Repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.Repo.FindByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT with claims matching middleware expectations
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(s.JWTSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
