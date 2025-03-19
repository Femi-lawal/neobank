package resilience

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name        string
	maxFailures int
	timeout     time.Duration
	halfOpenMax int

	mu            sync.RWMutex
	state         State
	failures      int
	successes     int
	lastFailure   time.Time
	halfOpenCount int
}

// CircuitBreakerConfig holds configuration for the circuit breaker
type CircuitBreakerConfig struct {
	Name             string
	MaxFailures      int           // Failures before opening
	Timeout          time.Duration // Time to wait before half-open
	HalfOpenMaxCalls int           // Max calls in half-open state
}

// DefaultConfig returns secure defaults
func DefaultConfig(name string) *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		Name:             name,
		MaxFailures:      5,
		Timeout:          30 * time.Second,
		HalfOpenMaxCalls: 3,
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		name:        config.Name,
		maxFailures: config.MaxFailures,
		timeout:     config.Timeout,
		halfOpenMax: config.HalfOpenMaxCalls,
		state:       StateClosed,
	}
}

// ErrCircuitOpen is returned when circuit is open
var ErrCircuitOpen = errors.New("circuit breaker is open")

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.canExecute() {
		return ErrCircuitOpen
	}

	err := fn()

	cb.recordResult(err)

	return err
}

// canExecute checks if the circuit allows requests
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = StateHalfOpen
			cb.halfOpenCount = 0
			return true
		}
		return false

	case StateHalfOpen:
		// Allow limited requests in half-open state
		if cb.halfOpenCount < cb.halfOpenMax {
			cb.halfOpenCount++
			return true
		}
		return false
	}

	return false
}

// recordResult records the result of an execution
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		cb.successes = 0

		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
	} else {
		cb.successes++

		if cb.state == StateHalfOpen {
			// Successful call in half-open state
			if cb.successes >= cb.halfOpenMax {
				cb.state = StateClosed
				cb.failures = 0
			}
		} else {
			cb.failures = 0
		}
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.halfOpenCount = 0
}

// CircuitBreakerRegistry manages multiple circuit breakers
type CircuitBreakerRegistry struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
}

// NewRegistry creates a new circuit breaker registry
func NewRegistry() *CircuitBreakerRegistry {
	return &CircuitBreakerRegistry{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// Get returns or creates a circuit breaker by name
func (r *CircuitBreakerRegistry) Get(name string) *CircuitBreaker {
	r.mu.RLock()
	cb, exists := r.breakers[name]
	r.mu.RUnlock()

	if exists {
		return cb
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, exists = r.breakers[name]; exists {
		return cb
	}

	cb = NewCircuitBreaker(DefaultConfig(name))
	r.breakers[name] = cb
	return cb
}

// Stats returns statistics for all circuit breakers
func (r *CircuitBreakerRegistry) Stats() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]string)
	for name, cb := range r.breakers {
		stats[name] = cb.State().String()
	}
	return stats
}
