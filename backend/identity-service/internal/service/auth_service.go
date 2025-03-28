package service

import (
	"errors"
	"time"

	"github.com/femi-lawal/new_bank/backend/identity-service/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// SEC-009: Use bcrypt cost factor 12 as documented in SECURITY.md
const BcryptCost = 12

// SEC-010: Token expiry times
const (
	AccessTokenExpiry  = 15 * time.Minute   // Short-lived access token
	RefreshTokenExpiry = 7 * 24 * time.Hour // Long-lived refresh token
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrAccountLocked      = errors.New("account is temporarily locked due to too many failed attempts")
)

type UserRepository interface {
	FindByEmail(email string) (*model.User, error)
	Create(user *model.User) error
}

type AuthService struct {
	Repo           UserRepository
	JWTSecret      []byte
	AccountLockout *AccountLockout // SEC-011: Account lockout integration
}

func NewAuthService(repo UserRepository, secret string) *AuthService {
	return &AuthService{
		Repo:           repo,
		JWTSecret:      []byte(secret),
		AccountLockout: DefaultAccountLockout(), // SEC-011: Initialize lockout
	}
}

func (s *AuthService) Register(email, password, firstName, lastName string) (*model.User, error) {
	// Check if user exists
	if _, err := s.Repo.FindByEmail(email); err == nil {
		return nil, ErrUserExists
	}

	// SEC-009: Use explicit bcrypt cost of 12
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
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
	// SEC-011: Check if account is locked
	if s.AccountLockout != nil && s.AccountLockout.IsLocked(email) {
		return "", ErrAccountLocked
	}

	user, err := s.Repo.FindByEmail(email)
	if err != nil {
		// SEC-011: Record failed attempt even for non-existent users (prevent enumeration)
		if s.AccountLockout != nil {
			s.AccountLockout.RecordFailedAttempt(email)
		}
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// SEC-011: Record failed attempt
		if s.AccountLockout != nil {
			s.AccountLockout.RecordFailedAttempt(email)
		}
		return "", ErrInvalidCredentials
	}

	// SEC-011: Clear failed attempts on successful login
	if s.AccountLockout != nil {
		s.AccountLockout.RecordSuccessfulLogin(email)
	}

	// SEC-010: Generate JWT with 15-minute expiry (was 24h)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(AccessTokenExpiry).Unix(),
	})

	tokenString, err := token.SignedString(s.JWTSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
