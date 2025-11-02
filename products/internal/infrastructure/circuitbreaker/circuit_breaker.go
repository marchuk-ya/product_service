package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type Config struct {
	MaxFailures      int
	Timeout          time.Duration
	SuccessThreshold int
}

func DefaultConfig() Config {
	return Config{
		MaxFailures:      5,
		Timeout:          60 * time.Second,
		SuccessThreshold: 2,
	}
}

type CircuitBreaker struct {
	mu            sync.RWMutex
	state         State
	failures      int
	successes     int
	lastFailure   time.Time
	config        Config
}

func NewCircuitBreaker(config Config) *CircuitBreaker {
	return &CircuitBreaker{
		state:  StateClosed,
		config: config,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.updateState()

	if cb.state == StateOpen {
		return errors.New("circuit breaker is open")
	}

	err := fn()

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

func (cb *CircuitBreaker) updateState() {
	switch cb.state {
	case StateOpen:
		if time.Since(cb.lastFailure) >= cb.config.Timeout {
			cb.state = StateHalfOpen
			cb.successes = 0
		}
	case StateHalfOpen:
	case StateClosed:
	}
}

func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateHalfOpen:
		cb.state = StateOpen
		cb.successes = 0
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			cb.state = StateOpen
		}
	}
}

func (cb *CircuitBreaker) onSuccess() {
	cb.failures = 0

	switch cb.state {
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			cb.state = StateClosed
			cb.successes = 0
		}
	case StateClosed:
		cb.failures = 0
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

