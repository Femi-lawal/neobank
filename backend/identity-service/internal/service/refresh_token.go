package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// RefreshTokenClaims represents claims for refresh token
type RefreshTokenClaims struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// RefreshToken represents a stored refresh token
type RefreshToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
	Revoked   bool
}

// GenerateTokenPair generates both access and refresh tokens
func (s *AuthService) GenerateTokenPair(userID string) (*TokenPair, error) {
	// Generate access token (short-lived)
	accessToken, err := s.generateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	// Generate refresh token (long-lived)
	refreshToken, err := s.generateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTokenExpiry.Seconds()),
	}, nil
}

// generateAccessToken creates a short-lived JWT access token
func (s *AuthService) generateAccessToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "neobank",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.JWTSecret)
}

// generateRefreshToken creates a long-lived refresh token
func (s *AuthService) generateRefreshToken(userID string) (string, error) {
	// Generate random token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	refreshToken := hex.EncodeToString(bytes)

	// In production, store refresh token in database
	// s.Repo.StoreRefreshToken(userID, refreshToken, time.Now().Add(7*24*time.Hour))

	return refreshToken, nil
}

// RefreshAccessToken validates refresh token and issues new access token
func (s *AuthService) RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	// In production, validate refresh token from database
	// rt, err := s.Repo.GetRefreshToken(refreshToken)
	// if err != nil || rt.Revoked || time.Now().After(rt.ExpiresAt) {
	//     return nil, errors.New("invalid or expired refresh token")
	// }

	// For demo, we'll validate the token format
	if len(refreshToken) != 64 {
		return nil, errors.New("invalid refresh token format")
	}

	// In production, get user ID from stored token
	// userID := rt.UserID

	// For now, return error as we need database integration
	return nil, errors.New("refresh token validation requires database integration")
}

// RevokeRefreshToken revokes a refresh token
func (s *AuthService) RevokeRefreshToken(refreshToken string) error {
	// In production:
	// return s.Repo.RevokeRefreshToken(refreshToken)
	return nil
}

// RevokeAllRefreshTokens revokes all refresh tokens for a user
func (s *AuthService) RevokeAllRefreshTokens(userID string) error {
	// In production:
	// return s.Repo.RevokeAllRefreshTokensForUser(userID)
	return nil
}
