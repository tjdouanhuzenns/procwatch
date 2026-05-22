package supervisor

import (
	"testing"
	"time"
)

func TestRetryPolicy_DefaultAllows(t *testing.T) {
	m := NewRetryPolicyManager()
	now := time.Now()
	if !m.Allow("svc", now) {
		t.Fatal("expected Allow to return true before any attempts")
	}
}

func TestRetryPolicy_BlocksAfterMaxAttempts(t *testing.T) {
	m := NewRetryPolicyManager()
	p := RetryPolicy{MaxAttempts: 3, Window: time.Minute, Cooldown: time.Minute}
	_ = m.SetPolicy("svc", p)

	now := time.Now()
	for i := 0; i < 3; i++ {
		m.Record("svc", now)
	}
	if m.Allow("svc", now) {
		t.Fatal("expected Allow to return false after max attempts")
	}
}

func TestRetryPolicy_AllowsBelowMax(t *testing.T) {
	m := NewRetryPolicyManager()
	p := RetryPolicy{MaxAttempts: 5, Window: time.Minute, Cooldown: time.Minute}
	_ = m.SetPolicy("svc", p)

	now := time.Now()
	for i := 0; i < 4; i++ {
		m.Record("svc", now)
	}
	if !m.Allow("svc", now) {
		t.Fatal("expected Allow to return true below max attempts")
	}
}

func TestRetryPolicy_OldAttemptsExpire(t *testing.T) {
	m := NewRetryPolicyManager()
	p := RetryPolicy{MaxAttempts: 2, Window: 10 * time.Second, Cooldown: time.Minute}
	_ = m.SetPolicy("svc", p)

	old := time.Now().Add(-20 * time.Second)
	m.Record("svc", old)
	m.Record("svc", old)

	now := time.Now()
	if !m.Allow("svc", now) {
		t.Fatal("expected old attempts to be outside window and Allow to return true")
	}
}

func TestRetryPolicy_CooldownBlocks(t *testing.T) {
	m := NewRetryPolicyManager()
	p := RetryPolicy{MaxAttempts: 1, Window: time.Minute, Cooldown: time.Hour}
	_ = m.SetPolicy("svc", p)

	now := time.Now()
	m.Record("svc", now)
	m.Allow("svc", now) // triggers cooldown

	if m.Allow("svc", now.Add(time.Second)) {
		t.Fatal("expected cooldown to block Allow")
	}
}

func TestRetryPolicy_ResetClearsCooldown(t *testing.T) {
	m := NewRetryPolicyManager()
	p := RetryPolicy{MaxAttempts: 1, Window: time.Minute, Cooldown: time.Hour}
	_ = m.SetPolicy("svc", p)

	now := time.Now()
	m.Record("svc", now)
	m.Allow("svc", now)
	m.Reset("svc")

	if !m.Allow("svc", now.Add(time.Second)) {
		t.Fatal("expected Allow to return true after Reset")
	}
}

func TestRetryPolicy_SetEmptyNameErrors(t *testing.T) {
	m := NewRetryPolicyManager()
	if err := m.SetPolicy("", DefaultRetryPolicy()); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRetryPolicy_GetReturnsDefault(t *testing.T) {
	m := NewRetryPolicyManager()
	p := m.GetPolicy("unknown")
	def := DefaultRetryPolicy()
	if p.MaxAttempts != def.MaxAttempts {
		t.Fatalf("expected default MaxAttempts %d, got %d", def.MaxAttempts, p.MaxAttempts)
	}
}

func TestRetryPolicy_IndependentPerProcess(t *testing.T) {
	m := NewRetryPolicyManager()
	p := RetryPolicy{MaxAttempts: 2, Window: time.Minute, Cooldown: time.Hour}
	_ = m.SetPolicy("a", p)
	_ = m.SetPolicy("b", p)

	now := time.Now()
	m.Record("a", now)
	m.Record("a", now)
	m.Allow("a", now) // trigger cooldown for a

	if !m.Allow("b", now) {
		t.Fatal("expected process b to be unaffected by a's cooldown")
	}
}
