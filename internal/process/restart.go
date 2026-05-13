package process

import "time"

// Policy defines the restart behavior for a supervised process.
type Policy string

const (
	// PolicyAlways restarts the process regardless of exit code.
	PolicyAlways Policy = "always"
	// PolicyOnFailure restarts only when the process exits with a non-zero code.
	PolicyOnFailure Policy = "on-failure"
	// PolicyNever never restarts the process.
	PolicyNever Policy = "never"
)

// RestartConfig holds configuration for the restart policy.
type RestartConfig struct {
	Policy      Policy
	MaxRetries  int
	BackoffBase time.Duration
	BackoffMax  time.Duration
}

// DefaultRestartConfig returns a RestartConfig with sensible defaults.
func DefaultRestartConfig() RestartConfig {
	return RestartConfig{
		Policy:      PolicyOnFailure,
		MaxRetries:  5,
		BackoffBase: 1 * time.Second,
		BackoffMax:  30 * time.Second,
	}
}

// RestartTracker tracks restart attempts and computes backoff delays.
type RestartTracker struct {
	cfg      RestartConfig
	attempts int
}

// NewRestartTracker creates a RestartTracker for the given config.
func NewRestartTracker(cfg RestartConfig) *RestartTracker {
	return &RestartTracker{cfg: cfg}
}

// ShouldRestart returns true if the process should be restarted given the
// exit code and current attempt count.
func (r *RestartTracker) ShouldRestart(exitCode int) bool {
	if r.cfg.MaxRetries >= 0 && r.attempts >= r.cfg.MaxRetries {
		return false
	}
	switch r.cfg.Policy {
	case PolicyAlways:
		return true
	case PolicyOnFailure:
		return exitCode != 0
	case PolicyNever:
		return false
	default:
		return false
	}
}

// NextBackoff increments the attempt counter and returns the delay before the
// next restart attempt using exponential backoff capped at BackoffMax.
func (r *RestartTracker) NextBackoff() time.Duration {
	r.attempts++
	delay := r.cfg.BackoffBase
	for i := 1; i < r.attempts; i++ {
		delay *= 2
		if delay > r.cfg.BackoffMax {
			delay = r.cfg.BackoffMax
			break
		}
	}
	return delay
}

// Attempts returns the number of restart attempts recorded so far.
func (r *RestartTracker) Attempts() int {
	return r.attempts
}

// Reset clears the attempt counter.
func (r *RestartTracker) Reset() {
	r.attempts = 0
}
