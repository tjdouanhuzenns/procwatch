package supervisor

import (
	"errors"
	"sync"
	"time"
)

// RetryPolicy defines per-process retry behaviour beyond the basic restart tracker.
type RetryPolicy struct {
	MaxAttempts int
	Window      time.Duration
	Cooldown    time.Duration
}

// DefaultRetryPolicy returns sensible defaults.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 5,
		Window:      2 * time.Minute,
		Cooldown:    30 * time.Second,
	}
}

// RetryPolicyManager tracks per-process attempt history and enforces policies.
type RetryPolicyManager struct {
	mu       sync.Mutex
	policies map[string]RetryPolicy
	attempts map[string][]time.Time
	cooling  map[string]time.Time
}

// NewRetryPolicyManager creates a new manager.
func NewRetryPolicyManager() *RetryPolicyManager {
	return &RetryPolicyManager{
		policies: make(map[string]RetryPolicy),
		attempts: make(map[string][]time.Time),
		cooling:  make(map[string]time.Time),
	}
}

// SetPolicy registers or replaces the policy for a named process.
func (m *RetryPolicyManager) SetPolicy(name string, p RetryPolicy) error {
	if name == "" {
		return errors.New("retrypolicy: name must not be empty")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.policies[name] = p
	return nil
}

// GetPolicy returns the policy for a process, or the default if none is set.
func (m *RetryPolicyManager) GetPolicy(name string) RetryPolicy {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.policies[name]; ok {
		return p
	}
	return DefaultRetryPolicy()
}

// Allow returns true if the process is permitted to restart now.
func (m *RetryPolicyManager) Allow(name string, now time.Time) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	p := m.policies[name]
	if p.MaxAttempts == 0 {
		p = DefaultRetryPolicy()
	}

	if until, ok := m.cooling[name]; ok && now.Before(until) {
		return false
	}

	windowStart := now.Add(-p.Window)
	var recent []time.Time
	for _, t := range m.attempts[name] {
		if t.After(windowStart) {
			recent = append(recent, t)
		}
	}
	m.attempts[name] = recent

	if len(recent) >= p.MaxAttempts {
		m.cooling[name] = now.Add(p.Cooldown)
		return false
	}
	return true
}

// Record registers a restart attempt for the named process.
func (m *RetryPolicyManager) Record(name string, at time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.attempts[name] = append(m.attempts[name], at)
}

// Reset clears all attempt history and cooldown for the named process.
func (m *RetryPolicyManager) Reset(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.attempts, name)
	delete(m.cooling, name)
}
