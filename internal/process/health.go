package process

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HealthCheck defines the configuration for a process health check.
type HealthCheck struct {
	URL      string        `toml:"url"`
	Interval time.Duration `toml:"interval"`
	Timeout  time.Duration `toml:"timeout"`
	Retries  int           `toml:"retries"`
}

// HealthChecker runs periodic HTTP health checks against a process.
type HealthChecker struct {
	cfg    HealthCheck
	client *http.Client
	onFail func(err error)
}

// DefaultHealthCheck returns a HealthCheck with sensible defaults.
func DefaultHealthCheck() HealthCheck {
	return HealthCheck{
		Interval: 10 * time.Second,
		Timeout:  2 * time.Second,
		Retries:  3,
	}
}

// NewHealthChecker creates a HealthChecker for the given config.
// onFail is called each time the health check fails after all retries.
func NewHealthChecker(cfg HealthCheck, onFail func(err error)) *HealthChecker {
	if cfg.Interval == 0 {
		cfg.Interval = DefaultHealthCheck().Interval
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultHealthCheck().Timeout
	}
	if cfg.Retries == 0 {
		cfg.Retries = DefaultHealthCheck().Retries
	}
	return &HealthChecker{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Timeout},
		onFail: onFail,
	}
}

// Run starts the health check loop, blocking until ctx is cancelled.
func (h *HealthChecker) Run(ctx context.Context) {
	ticker := time.NewTicker(h.cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := h.check(ctx); err != nil {
				h.onFail(err)
			}
		}
	}
}

// check performs the HTTP GET with retries.
func (h *HealthChecker) check(ctx context.Context) error {
	var lastErr error
	for i := 0; i < h.cfg.Retries; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.cfg.URL, nil)
		if err != nil {
			return fmt.Errorf("health check request build failed: %w", err)
		}
		resp, err := h.client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return nil
			}
			lastErr = fmt.Errorf("health check got status %d", resp.StatusCode)
		} else {
			lastErr = fmt.Errorf("health check request failed: %w", err)
		}
	}
	return lastErr
}
