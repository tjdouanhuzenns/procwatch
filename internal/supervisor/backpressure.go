package supervisor

import (
	"context"
	"sync"
	"time"
)

// BackpressureConfig controls how the backpressure valve behaves.
type BackpressureConfig struct {
	// MaxPending is the maximum number of pending restart requests before backpressure is applied.
	MaxPending int
	// WaitTimeout is how long a caller will wait for capacity before giving up.
	WaitTimeout time.Duration
}

// DefaultBackpressureConfig returns sensible defaults.
func DefaultBackpressureConfig() BackpressureConfig {
	return BackpressureConfig{
		MaxPending:  10,
		WaitTimeout: 2 * time.Second,
	}
}

// Backpressure limits the rate of concurrent restart requests accepted by the supervisor.
type Backpressure struct {
	cfg BackpressureConfig
	sem chan struct{}
	mu  sync.Mutex
}

// NewBackpressure creates a new Backpressure valve with the given config.
func NewBackpressure(cfg BackpressureConfig) *Backpressure {
	return &Backpressure{
		cfg: cfg,
		sem: make(chan struct{}, cfg.MaxPending),
	}
}

// Acquire attempts to acquire a slot within the configured wait timeout.
// Returns false if no slot is available before the deadline or context is cancelled.
func (b *Backpressure) Acquire(ctx context.Context) bool {
	deadline := time.NewTimer(b.cfg.WaitTimeout)
	defer deadline.Stop()
	select {
	case b.sem <- struct{}{}:
		return true
	case <-deadline.C:
		return false
	case <-ctx.Done():
		return false
	}
}

// Release frees a previously acquired slot.
func (b *Backpressure) Release() {
	select {
	case <-b.sem:
	default:
	}
}

// Pending returns the number of currently held slots.
func (b *Backpressure) Pending() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.sem)
}

// Reset drains all held slots.
func (b *Backpressure) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for {
		select {
		case <-b.sem:
		default:
			return
		}
	}
}
