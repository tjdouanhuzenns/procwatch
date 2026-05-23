package supervisor

import (
	"testing"
	"time"
)

func TestProcessCooldown_NotCoolingDownInitially(t *testing.T) {
	c := NewProcessCooldown(5 * time.Second)
	if c.IsCoolingDown("svc") {
		t.Fatal("expected process to not be cooling down initially")
	}
}

func TestProcessCooldown_SetAndCheck(t *testing.T) {
	c := NewProcessCooldown(5 * time.Second)
	if err := c.Set("svc", 200*time.Millisecond); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.IsCoolingDown("svc") {
		t.Fatal("expected process to be cooling down after Set")
	}
}

func TestProcessCooldown_ExpiresAfterDuration(t *testing.T) {
	c := NewProcessCooldown(5 * time.Second)
	_ = c.Set("svc", 30*time.Millisecond)
	time.Sleep(60 * time.Millisecond)
	if c.IsCoolingDown("svc") {
		t.Fatal("expected cooldown to have expired")
	}
}

func TestProcessCooldown_Clear(t *testing.T) {
	c := NewProcessCooldown(5 * time.Second)
	_ = c.Set("svc", 10*time.Second)
	c.Clear("svc")
	if c.IsCoolingDown("svc") {
		t.Fatal("expected cooldown to be cleared")
	}
}

func TestProcessCooldown_Remaining(t *testing.T) {
	c := NewProcessCooldown(5 * time.Second)
	_ = c.Set("svc", 500*time.Millisecond)
	rem := c.Remaining("svc")
	if rem <= 0 || rem > 500*time.Millisecond {
		t.Fatalf("unexpected remaining duration: %v", rem)
	}
}

func TestProcessCooldown_RemainingZeroWhenNotSet(t *testing.T) {
	c := NewProcessCooldown(5 * time.Second)
	if c.Remaining("missing") != 0 {
		t.Fatal("expected zero remaining for unknown process")
	}
}

func TestProcessCooldown_EmptyNameErrors(t *testing.T) {
	c := NewProcessCooldown(5 * time.Second)
	if err := c.Set("", time.Second); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestProcessCooldown_UsesDefaultTTL(t *testing.T) {
	c := NewProcessCooldown(100 * time.Millisecond)
	_ = c.Set("svc", 0) // zero triggers default
	if !c.IsCoolingDown("svc") {
		t.Fatal("expected cooldown to be active with default TTL")
	}
	time.Sleep(150 * time.Millisecond)
	if c.IsCoolingDown("svc") {
		t.Fatal("expected default TTL cooldown to have expired")
	}
}

func TestProcessCooldown_All(t *testing.T) {
	c := NewProcessCooldown(5 * time.Second)
	_ = c.Set("a", time.Second)
	_ = c.Set("b", time.Second)
	all := c.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if _, ok := all["a"]; !ok {
		t.Error("expected entry for 'a'")
	}
	if _, ok := all["b"]; !ok {
		t.Error("expected entry for 'b'")
	}
}
