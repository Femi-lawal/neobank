package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	Token     string
	UserID    string
	Email     string
	ExpiresAt time.Time
	Used      bool
}

// RequestPasswordReset initiates password reset process
func (s *AuthService) RequestPasswordReset(email string) (*PasswordResetToken, error) {
	// Find user by email
	user, err := s.Repo.FindByEmail(email)
	if err != nil {
		// Don't reveal if email exists
		return nil, nil
	}

	// Generate reset token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(bytes)

	resetToken := &PasswordResetToken{
		Token:     token,
		UserID:    user.ID.String(),
		Email:     email,
		ExpiresAt: time.Now().Add(1 * time.Hour), // Token valid for 1 hour
		Used:      false,
	}

	// In production, store token and send email
	// s.Repo.StorePasswordResetToken(resetToken)
	// s.EmailService.SendPasswordResetEmail(email, token)

	return resetToken, nil
}

// ValidatePasswordResetToken validates a password reset token
func (s *AuthService) ValidatePasswordResetToken(token string) (*PasswordResetToken, error) {
	// In production, get from database
	// resetToken, err := s.Repo.GetPasswordResetToken(token)
	// if err != nil {
	//     return nil, errors.New("invalid reset token")
	// }
	// if resetToken.Used {
	//     return nil, errors.New("token already used")
	// }
	// if time.Now().After(resetToken.ExpiresAt) {
	//     return nil, errors.New("token expired")
	// }
	// return resetToken, nil

	if len(token) != 64 {
		return nil, errors.New("invalid token format")
	}

	return nil, errors.New("token validation requires database integration")
}

// ResetPassword resets user password with valid token
func (s *AuthService) ResetPassword(token string, newPassword string) error {
	// Validate token
	resetToken, err := s.ValidatePasswordResetToken(token)
	if err != nil {
		return err
	}

	// Validate password strength
	if len(newPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	// Hash new password
	hashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update user password
	err = s.Repo.UpdatePassword(resetToken.UserID, hashedPassword)
	if err != nil {
		return err
	}

	// Mark token as used
	// s.Repo.MarkPasswordResetTokenUsed(token)

	// Revoke all refresh tokens for security
	s.RevokeAllRefreshTokens(resetToken.UserID)

	return nil
}

// ChangePassword changes password for authenticated user
func (s *AuthService) ChangePassword(userID string, currentPassword string, newPassword string) error {
	// Get user
	user, err := s.Repo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify current password
	if err := s.verifyPassword(user.PasswordHash, currentPassword); err != nil {
		return errors.New("current password is incorrect")
	}

	// Validate new password
	if len(newPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}

	if currentPassword == newPassword {
		return errors.New("new password must be different from current password")
	}

	// Hash new password
	hashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	return s.Repo.UpdatePassword(userID, hashedPassword)
}
