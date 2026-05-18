package supervisor

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig holds configuration for the circuit breaker.
type CircuitBreakerConfig struct {
	FailureThreshold int
	SuccessThreshold int
	OpenDuration     time.Duration
}

// DefaultCircuitBreakerConfig returns sensible defaults.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		OpenDuration:     30 * time.Second,
	}
}

// CircuitBreaker tracks failure rates per process and trips open when exceeded.
type CircuitBreaker struct {
	mu       sync.Mutex
	cfg      CircuitBreakerConfig
	states   map[string]*circuitEntry
}

type circuitEntry struct {
	state    CircuitState
	failures int
	successes int
	openedAt time.Time
}

// NewCircuitBreaker creates a new CircuitBreaker with the given config.
func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		cfg:    cfg,
		states: make(map[string]*circuitEntry),
	}
}

func (cb *CircuitBreaker) entry(name string) *circuitEntry {
	if e, ok := cb.states[name]; ok {
		return e
	}
	e := &circuitEntry{state: CircuitClosed}
	cb.states[name] = e
	return e
}

// Allow returns true if the process is permitted to start.
func (cb *CircuitBreaker) Allow(name string) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	e := cb.entry(name)
	if e.state == CircuitOpen {
		if time.Since(e.openedAt) >= cb.cfg.OpenDuration {
			e.state = CircuitHalfOpen
			e.successes = 0
			return true
		}
		return false
	}
	return true
}

// RecordSuccess records a successful run for the named process.
func (cb *CircuitBreaker) RecordSuccess(name string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	e := cb.entry(name)
	if e.state == CircuitHalfOpen {
		e.successes++
		if e.successes >= cb.cfg.SuccessThreshold {
			e.state = CircuitClosed
			e.failures = 0
		}
	} else {
		e.failures = 0
	}
}

// RecordFailure records a failed run for the named process.
func (cb *CircuitBreaker) RecordFailure(name string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	e := cb.entry(name)
	e.failures++
	if e.state != CircuitOpen && e.failures >= cb.cfg.FailureThreshold {
		e.state = CircuitOpen
		e.openedAt = time.Now()
	}
}

// State returns the current circuit state for a process.
func (cb *CircuitBreaker) State(name string) CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.entry(name).state
}

// Reset clears circuit state for a process.
func (cb *CircuitBreaker) Reset(name string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.states[name] = &circuitEntry{state: CircuitClosed}
}
