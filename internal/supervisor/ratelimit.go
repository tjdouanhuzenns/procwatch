package supervisor

import (
	"context"
	"sync"
	"time"
)

// RateLimitConfig controls how many restart attempts are allowed per process
// within a rolling time window before the process is considered crashed.
type RateLimitConfig struct {
	MaxRestarts int
	Window      time.Duration
	Cooldown    time.Duration
}

// DefaultRateLimitConfig returns a sensible default rate limit configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxRestarts: 5,
		Window:      time.Minute,
		Cooldown:    30 * time.Second,
	}
}

// RateLimiter tracks restart attempts per process and enforces a rate limit.
type RateLimiter struct {
	mu       sync.Mutex
	cfg      RateLimitConfig
	events   map[string][]time.Time
	cooling  map[string]time.Time
}

// NewRateLimiter creates a new RateLimiter with the given configuration.
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		cfg:     cfg,
		events:  make(map[string][]time.Time),
		cooling: make(map[string]time.Time),
	}
}

// Allow returns true if the process with the given name is permitted to restart.
// It records the attempt and prunes stale events outside the window.
func (r *RateLimiter) Allow(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	if until, ok := r.cooling[name]; ok && now.Before(until) {
		return false
	}

	cutoff := now.Add(-r.cfg.Window)
	evts := r.events[name]
	filtered := evts[:0]
	for _, t := range evts {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= r.cfg.MaxRestarts {
		r.cooling[name] = now.Add(r.cfg.Cooldown)
		r.events[name] = filtered[:0]
		return false
	}

	r.events[name] = append(filtered, now)
	return true
}

// Reset clears all state for a process (e.g. after a successful healthy run).
func (r *RateLimiter) Reset(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.events, name)
	delete(r.cooling, name)
}

// CoolingUntil returns the time until which the named process is rate-limited,
// and false if it is not currently cooling down.
func (r *RateLimiter) CoolingUntil(name string) (time.Time, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	until, ok := r.cooling[name]
	if !ok || time.Now().After(until) {
		return time.Time{}, false
	}
	return until, true
}

// RunWithContext is a helper that blocks until ctx is done; it exists to satisfy
// supervisor lifecycle conventions and can be embedded in a supervisor loop.
func (r *RateLimiter) RunWithContext(ctx context.Context) {
	<-ctx.Done()
}
