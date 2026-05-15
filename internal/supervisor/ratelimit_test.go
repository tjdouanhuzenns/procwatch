package supervisor

import (
	"testing"
	"time"
)

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	cfg := RateLimitConfig{MaxRestarts: 3, Window: time.Minute, Cooldown: 10 * time.Second}
	rl := NewRateLimiter(cfg)

	for i := 0; i < 3; i++ {
		if !rl.Allow("svc") {
			t.Fatalf("expected Allow=true on attempt %d", i+1)
		}
	}
}

func TestRateLimiter_BlocksAtLimit(t *testing.T) {
	cfg := RateLimitConfig{MaxRestarts: 3, Window: time.Minute, Cooldown: 10 * time.Second}
	rl := NewRateLimiter(cfg)

	for i := 0; i < 3; i++ {
		rl.Allow("svc")
	}
	if rl.Allow("svc") {
		t.Fatal("expected Allow=false after reaching max restarts")
	}
}

func TestRateLimiter_CoolingDownAfterLimit(t *testing.T) {
	cfg := RateLimitConfig{MaxRestarts: 2, Window: time.Minute, Cooldown: 5 * time.Second}
	rl := NewRateLimiter(cfg)

	rl.Allow("svc")
	rl.Allow("svc")
	rl.Allow("svc") // triggers cooldown

	until, ok := rl.CoolingUntil("svc")
	if !ok {
		t.Fatal("expected process to be in cooldown")
	}
	if time.Until(until) > 6*time.Second {
		t.Errorf("cooldown duration unexpectedly long: %v", time.Until(until))
	}
}

func TestRateLimiter_ResetClearsCooldown(t *testing.T) {
	cfg := RateLimitConfig{MaxRestarts: 1, Window: time.Minute, Cooldown: time.Minute}
	rl := NewRateLimiter(cfg)

	rl.Allow("svc")
	rl.Allow("svc") // triggers cooldown

	rl.Reset("svc")

	if !rl.Allow("svc") {
		t.Fatal("expected Allow=true after reset")
	}
	if _, ok := rl.CoolingUntil("svc"); ok {
		t.Fatal("expected no cooldown after reset")
	}
}

func TestRateLimiter_IndependentPerProcess(t *testing.T) {
	cfg := RateLimitConfig{MaxRestarts: 2, Window: time.Minute, Cooldown: time.Minute}
	rl := NewRateLimiter(cfg)

	rl.Allow("a")
	rl.Allow("a")
	rl.Allow("a") // a is now cooling

	if !rl.Allow("b") {
		t.Fatal("process b should not be affected by process a's rate limit")
	}
}

func TestRateLimiter_OldEventsExpire(t *testing.T) {
	cfg := RateLimitConfig{MaxRestarts: 2, Window: 50 * time.Millisecond, Cooldown: time.Second}
	rl := NewRateLimiter(cfg)

	rl.Allow("svc")
	rl.Allow("svc")

	time.Sleep(60 * time.Millisecond)

	// Old events should have expired; new window starts fresh.
	if !rl.Allow("svc") {
		t.Fatal("expected Allow=true after window expiry")
	}
}
