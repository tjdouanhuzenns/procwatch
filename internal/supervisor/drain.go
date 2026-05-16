package supervisor

import (
	"context"
	"sync"
	"time"
)

// DrainConfig holds configuration for graceful drain behaviour.
type DrainConfig struct {
	// Timeout is the maximum time to wait for processes to finish.
	Timeout time.Duration
	// PollInterval is how often to check if all processes are done.
	PollInterval time.Duration
}

// DefaultDrainConfig returns sensible drain defaults.
func DefaultDrainConfig() DrainConfig {
	return DrainConfig{
		Timeout:      30 * time.Second,
		PollInterval: 250 * time.Millisecond,
	}
}

// Drainer coordinates a graceful shutdown, waiting for tracked
// processes to reach a terminal state before the deadline.
type Drainer struct {
	mu      sync.Mutex
	active  map[string]struct{}
	cfg     DrainConfig
	done    chan struct{}
}

// NewDrainer creates a Drainer with the given config.
func NewDrainer(cfg DrainConfig) *Drainer {
	return &Drainer{
		active: make(map[string]struct{}),
		cfg:    cfg,
		done:   make(chan struct{}),
	}
}

// Track registers a process name as active.
func (d *Drainer) Track(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.active[name] = struct{}{}
}

// Release marks a process as finished.
func (d *Drainer) Release(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.active, name)
	if len(d.active) == 0 {
		select {
		case <-d.done:
		default:
			close(d.done)
		}
	}
}

// ActiveCount returns the number of processes still running.
func (d *Drainer) ActiveCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.active)
}

// Wait blocks until all processes are released or the context deadline
// is exceeded. Returns true if all processes drained cleanly.
func (d *Drainer) Wait(ctx context.Context) bool {
	timeout := time.NewTimer(d.cfg.Timeout)
	defer timeout.Stop()
	for {
		select {
		case <-d.done:
			return true
		case <-ctx.Done():
			return false
		case <-timeout.C:
			return false
		case <-time.After(d.cfg.PollInterval):
			if d.ActiveCount() == 0 {
				return true
			}
		}
	}
}
