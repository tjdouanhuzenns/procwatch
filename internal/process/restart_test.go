package process

import (
	"testing"
	"time"
)

func TestShouldRestart_Always(t *testing.T) {
	cfg := RestartConfig{Policy: PolicyAlways, MaxRetries: 3, BackoffBase: time.Second, BackoffMax: 10 * time.Second}
	rt := NewRestartTracker(cfg)
	if !rt.ShouldRestart(0) {
		t.Error("expected restart on exit 0 with PolicyAlways")
	}
	if !rt.ShouldRestart(1) {
		t.Error("expected restart on exit 1 with PolicyAlways")
	}
}

func TestShouldRestart_OnFailure(t *testing.T) {
	cfg := RestartConfig{Policy: PolicyOnFailure, MaxRetries: 3, BackoffBase: time.Second, BackoffMax: 10 * time.Second}
	rt := NewRestartTracker(cfg)
	if rt.ShouldRestart(0) {
		t.Error("expected no restart on exit 0 with PolicyOnFailure")
	}
	if !rt.ShouldRestart(1) {
		t.Error("expected restart on exit 1 with PolicyOnFailure")
	}
}

func TestShouldRestart_Never(t *testing.T) {
	cfg := RestartConfig{Policy: PolicyNever, MaxRetries: 3, BackoffBase: time.Second, BackoffMax: 10 * time.Second}
	rt := NewRestartTracker(cfg)
	if rt.ShouldRestart(1) {
		t.Error("expected no restart with PolicyNever")
	}
}

func TestShouldRestart_MaxRetries(t *testing.T) {
	cfg := RestartConfig{Policy: PolicyAlways, MaxRetries: 2, BackoffBase: time.Second, BackoffMax: 10 * time.Second}
	rt := NewRestartTracker(cfg)
	rt.NextBackoff()
	rt.NextBackoff()
	if rt.ShouldRestart(1) {
		t.Error("expected no restart after MaxRetries exceeded")
	}
}

func TestNextBackoff_ExponentialGrowth(t *testing.T) {
	cfg := RestartConfig{Policy: PolicyAlways, MaxRetries: 10, BackoffBase: 1 * time.Second, BackoffMax: 30 * time.Second}
	rt := NewRestartTracker(cfg)

	d1 := rt.NextBackoff() // attempts=1 → 1s
	d2 := rt.NextBackoff() // attempts=2 → 2s
	d3 := rt.NextBackoff() // attempts=3 → 4s

	if d1 != time.Second {
		t.Errorf("expected 1s, got %v", d1)
	}
	if d2 != 2*time.Second {
		t.Errorf("expected 2s, got %v", d2)
	}
	if d3 != 4*time.Second {
		t.Errorf("expected 4s, got %v", d3)
	}
}

func TestNextBackoff_CappedAtMax(t *testing.T) {
	cfg := RestartConfig{Policy: PolicyAlways, MaxRetries: 20, BackoffBase: 1 * time.Second, BackoffMax: 4 * time.Second}
	rt := NewRestartTracker(cfg)
	for i := 0; i < 10; i++ {
		d := rt.NextBackoff()
		if d > cfg.BackoffMax {
			t.Errorf("backoff %v exceeded max %v on attempt %d", d, cfg.BackoffMax, i+1)
		}
	}
}

func TestReset(t *testing.T) {
	cfg := DefaultRestartConfig()
	rt := NewRestartTracker(cfg)
	rt.NextBackoff()
	rt.NextBackoff()
	rt.Reset()
	if rt.Attempts() != 0 {
		t.Errorf("expected 0 attempts after reset, got %d", rt.Attempts())
	}
}
