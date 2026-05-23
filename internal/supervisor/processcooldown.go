package supervisor

import (
	"fmt"
	"sync"
	"time"
)

// ProcessCooldown tracks per-process cooldown periods after crashes or restarts.
// It prevents a process from being immediately restarted while in a cooldown window.
type ProcessCooldown struct {
	mu      sync.RWMutex
	cooldowns map[string]time.Time
	defaultTTL time.Duration
}

// NewProcessCooldown creates a new ProcessCooldown with the given default cooldown duration.
func NewProcessCooldown(defaultTTL time.Duration) *ProcessCooldown {
	if defaultTTL <= 0 {
		defaultTTL = 5 * time.Second
	}
	return &ProcessCooldown{
		cooldowns:  make(map[string]time.Time),
		defaultTTL: defaultTTL,
	}
}

// Set marks the named process as cooling down for the given duration.
// If duration is zero, the default TTL is used.
func (c *ProcessCooldown) Set(name string, duration time.Duration) error {
	if name == "" {
		return fmt.Errorf("process name must not be empty")
	}
	if duration <= 0 {
		duration = c.defaultTTL
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cooldowns[name] = time.Now().Add(duration)
	return nil
}

// IsCoolingDown reports whether the named process is currently in a cooldown period.
func (c *ProcessCooldown) IsCoolingDown(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	expiry, ok := c.cooldowns[name]
	if !ok {
		return false
	}
	return time.Now().Before(expiry)
}

// Clear removes the cooldown entry for the named process.
func (c *ProcessCooldown) Clear(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cooldowns, name)
}

// Remaining returns the duration left in the cooldown for the named process.
// Returns zero if the process is not cooling down.
func (c *ProcessCooldown) Remaining(name string) time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	expiry, ok := c.cooldowns[name]
	if !ok {
		return 0
	}
	rem := time.Until(expiry)
	if rem < 0 {
		return 0
	}
	return rem
}

// All returns a map of process names to their cooldown expiry times.
func (c *ProcessCooldown) All() map[string]time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]time.Time, len(c.cooldowns))
	for k, v := range c.cooldowns {
		out[k] = v
	}
	return out
}
