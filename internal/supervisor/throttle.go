package supervisor

import (
	"sync"
	"time"
)

// ThrottleConfig holds configuration for restart throttling.
type ThrottleConfig struct {
	Window    time.Duration
	MaxStarts int
}

// DefaultThrottleConfig returns a sensible default throttle configuration.
func DefaultThrottleConfig() ThrottleConfig {
	return ThrottleConfig{
		Window:    30 * time.Second,
		MaxStarts: 5,
	}
}

// Throttle tracks process start events per process name and determines
// whether a process is being restarted too frequently.
type Throttle struct {
	mu     sync.Mutex
	cfg    ThrottleConfig
	events map[string][]time.Time
	now    func() time.Time
}

// NewThrottle creates a new Throttle with the given configuration.
func NewThrottle(cfg ThrottleConfig) *Throttle {
	return &Throttle{
		cfg:    cfg,
		events: make(map[string][]time.Time),
		now:    time.Now,
	}
}

// Record registers a start event for the named process.
func (t *Throttle) Record(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.now()
	t.events[name] = append(t.prune(name, now), now)
}

// IsThrottled returns true if the process has exceeded the allowed start
// count within the configured time window.
func (t *Throttle) IsThrottled(name string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.now()
	recent := t.prune(name, now)
	t.events[name] = recent
	return len(recent) >= t.cfg.MaxStarts
}

// Reset clears the start history for the named process.
func (t *Throttle) Reset(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.events, name)
}

// prune removes events outside the current window. Must be called with lock held.
func (t *Throttle) prune(name string, now time.Time) []time.Time {
	cutoff := now.Add(-t.cfg.Window)
	evts := t.events[name]
	var recent []time.Time
	for _, e := range evts {
		if e.After(cutoff) {
			recent = append(recent, e)
		}
	}
	return recent
}
