package service

import (
	"errors"
	"sync"
	"time"
)

// AccountLockout manages failed login attempts and account lockouts
type AccountLockout struct {
	attempts     map[string]*lockoutInfo
	mu           sync.RWMutex
	maxAttempts  int
	lockDuration time.Duration
	windowSize   time.Duration
}

type lockoutInfo struct {
	attempts     int
	firstAttempt time.Time
	lockedUntil  time.Time
}

// NewAccountLockout creates a new account lockout manager
func NewAccountLockout(maxAttempts int, lockDuration, windowSize time.Duration) *AccountLockout {
	return &AccountLockout{
		attempts:     make(map[string]*lockoutInfo),
		maxAttempts:  maxAttempts,
		lockDuration: lockDuration,
		windowSize:   windowSize,
	}
}

// DefaultAccountLockout creates a lockout with secure defaults
func DefaultAccountLockout() *AccountLockout {
	return NewAccountLockout(
		5,              // 5 failed attempts
		15*time.Minute, // 15 minute lockout
		10*time.Minute, // 10 minute window
	)
}

// IsLocked checks if an account is currently locked
func (al *AccountLockout) IsLocked(identifier string) bool {
	al.mu.RLock()
	defer al.mu.RUnlock()

	info, exists := al.attempts[identifier]
	if !exists {
		return false
	}

	return time.Now().Before(info.lockedUntil)
}

// RecordFailedAttempt records a failed login attempt
func (al *AccountLockout) RecordFailedAttempt(identifier string) (locked bool, remainingAttempts int) {
	al.mu.Lock()
	defer al.mu.Unlock()

	now := time.Now()
	info, exists := al.attempts[identifier]

	if !exists {
		info = &lockoutInfo{
			attempts:     1,
			firstAttempt: now,
		}
		al.attempts[identifier] = info
		return false, al.maxAttempts - 1
	}

	// Check if still locked
	if now.Before(info.lockedUntil) {
		return true, 0
	}

	// Reset if window expired
	if now.Sub(info.firstAttempt) > al.windowSize {
		info.attempts = 1
		info.firstAttempt = now
		info.lockedUntil = time.Time{}
		return false, al.maxAttempts - 1
	}

	// Increment attempts
	info.attempts++

	// Check if should lock
	if info.attempts >= al.maxAttempts {
		info.lockedUntil = now.Add(al.lockDuration)
		return true, 0
	}

	return false, al.maxAttempts - info.attempts
}

// RecordSuccessfulLogin clears failed attempts for an account
func (al *AccountLockout) RecordSuccessfulLogin(identifier string) {
	al.mu.Lock()
	defer al.mu.Unlock()

	delete(al.attempts, identifier)
}

// GetLockoutInfo returns lockout information for an account
func (al *AccountLockout) GetLockoutInfo(identifier string) (attempts int, lockedUntil time.Time, err error) {
	al.mu.RLock()
	defer al.mu.RUnlock()

	info, exists := al.attempts[identifier]
	if !exists {
		return 0, time.Time{}, nil
	}

	return info.attempts, info.lockedUntil, nil
}

// ErrAccountLocked is returned when an account is locked
var ErrAccountLocked = errors.New("account is temporarily locked due to too many failed attempts")

// Cleanup removes expired entries to prevent memory leaks
func (al *AccountLockout) Cleanup() {
	al.mu.Lock()
	defer al.mu.Unlock()

	now := time.Now()
	for identifier, info := range al.attempts {
		// Remove if lockout expired and window passed
		if now.After(info.lockedUntil) && now.Sub(info.firstAttempt) > al.windowSize {
			delete(al.attempts, identifier)
		}
	}
}

// StartCleanupRoutine starts a background goroutine to clean up expired entries
func (al *AccountLockout) StartCleanupRoutine(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			al.Cleanup()
		}
	}()
}
