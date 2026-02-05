package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreakerManager manages circuit breakers for multiple providers
type CircuitBreakerManager struct {
	breakers map[string]*gobreaker.CircuitBreaker
	mu       sync.RWMutex
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
	}
}

// GetBreaker returns or creates a circuit breaker for a provider
func (m *CircuitBreakerManager) GetBreaker(providerID string) *gobreaker.CircuitBreaker {
	m.mu.RLock()
	breaker, exists := m.breakers[providerID]
	m.mu.RUnlock()

	if exists {
		return breaker
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if breaker, exists := m.breakers[providerID]; exists {
		return breaker
	}

	// Create new circuit breaker
	settings := gobreaker.Settings{
		Name:        providerID,
		MaxRequests: 1,                // Max requests in half-open state
		Interval:    60 * time.Second, // Reset failure count after this interval
		Timeout:     60 * time.Second, // Stay open for this duration
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Open circuit after 5 consecutive failures
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// Log state changes
			fmt.Printf("Circuit breaker %s: %s -> %s\n", name, from, to)
		},
	}

	breaker = gobreaker.NewCircuitBreaker(settings)
	m.breakers[providerID] = breaker

	return breaker
}

// Execute executes a function with circuit breaker protection
func (m *CircuitBreakerManager) Execute(providerID string, fn func() (interface{}, error)) (interface{}, error) {
	breaker := m.GetBreaker(providerID)
	return breaker.Execute(fn)
}

// ExecuteWithFallback executes a function with circuit breaker and fallback
func (m *CircuitBreakerManager) ExecuteWithFallback(
	ctx context.Context,
	primaryProviderID string,
	primary func() (interface{}, error),
	fallbackProviderID string,
	fallback func() (interface{}, error),
) (interface{}, error) {
	// Try primary
	result, err := m.Execute(primaryProviderID, primary)
	if err == nil {
		return result, nil
	}

	// Primary failed, try fallback
	if fallback != nil && fallbackProviderID != "" {
		fmt.Printf("Primary provider %s failed, trying fallback %s\n", primaryProviderID, fallbackProviderID)
		result, fallbackErr := m.Execute(fallbackProviderID, fallback)
		if fallbackErr == nil {
			return result, nil
		}
		// Both failed
		return nil, fmt.Errorf("primary error: %w, fallback error: %v", err, fallbackErr)
	}

	return nil, err
}

// GetState returns the current state of a circuit breaker
func (m *CircuitBreakerManager) GetState(providerID string) gobreaker.State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if breaker, exists := m.breakers[providerID]; exists {
		return breaker.State()
	}

	return gobreaker.StateClosed // Default to closed if not found
}
