package supervisor

import (
	"testing"
	"time"
)

func testCBConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		OpenDuration:     100 * time.Millisecond,
	}
}

func TestCircuitBreaker_ClosedByDefault(t *testing.T) {
	cb := NewCircuitBreaker(testCBConfig())
	if !cb.Allow("svc") {
		t.Fatal("expected circuit to allow initially")
	}
	if cb.State("svc") != CircuitClosed {
		t.Fatalf("expected closed, got %s", cb.State("svc"))
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(testCBConfig())
	for i := 0; i < 3; i++ {
		cb.RecordFailure("svc")
	}
	if cb.State("svc") != CircuitOpen {
		t.Fatalf("expected open, got %s", cb.State("svc"))
	}
	if cb.Allow("svc") {
		t.Fatal("expected circuit to block when open")
	}
}

func TestCircuitBreaker_HalfOpenAfterDuration(t *testing.T) {
	cb := NewCircuitBreaker(testCBConfig())
	for i := 0; i < 3; i++ {
		cb.RecordFailure("svc")
	}
	time.Sleep(120 * time.Millisecond)
	if !cb.Allow("svc") {
		t.Fatal("expected half-open to allow one attempt")
	}
	if cb.State("svc") != CircuitHalfOpen {
		t.Fatalf("expected half-open, got %s", cb.State("svc"))
	}
}

func TestCircuitBreaker_ClosesAfterSuccessThreshold(t *testing.T) {
	cb := NewCircuitBreaker(testCBConfig())
	for i := 0; i < 3; i++ {
		cb.RecordFailure("svc")
	}
	time.Sleep(120 * time.Millisecond)
	cb.Allow("svc") // transitions to half-open
	cb.RecordSuccess("svc")
	cb.RecordSuccess("svc")
	if cb.State("svc") != CircuitClosed {
		t.Fatalf("expected closed after successes, got %s", cb.State("svc"))
	}
	if !cb.Allow("svc") {
		t.Fatal("expected circuit to allow after closing")
	}
}

func TestCircuitBreaker_SuccessResetFailures(t *testing.T) {
	cb := NewCircuitBreaker(testCBConfig())
	cb.RecordFailure("svc")
	cb.RecordFailure("svc")
	cb.RecordSuccess("svc")
	if cb.State("svc") != CircuitClosed {
		t.Fatalf("expected closed after success reset, got %s", cb.State("svc"))
	}
	// Should not open after only 2 more failures since counter was reset
	cb.RecordFailure("svc")
	cb.RecordFailure("svc")
	if cb.State("svc") != CircuitClosed {
		t.Fatalf("expected still closed, got %s", cb.State("svc"))
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(testCBConfig())
	for i := 0; i < 3; i++ {
		cb.RecordFailure("svc")
	}
	cb.Reset("svc")
	if cb.State("svc") != CircuitClosed {
		t.Fatalf("expected closed after reset, got %s", cb.State("svc"))
	}
	if !cb.Allow("svc") {
		t.Fatal("expected allow after reset")
	}
}

func TestCircuitBreaker_IndependentPerProcess(t *testing.T) {
	cb := NewCircuitBreaker(testCBConfig())
	for i := 0; i < 3; i++ {
		cb.RecordFailure("svc-a")
	}
	if cb.State("svc-b") != CircuitClosed {
		t.Fatal("svc-b should be unaffected by svc-a failures")
	}
	if !cb.Allow("svc-b") {
		t.Fatal("svc-b should still be allowed")
	}
}

func TestCircuitStateString(t *testing.T) {
	cases := map[CircuitState]string{
		CircuitClosed:   "closed",
		CircuitOpen:     "open",
		CircuitHalfOpen: "half-open",
		CircuitState(99): "unknown",
	}
	for state, want := range cases {
		if got := state.String(); got != want {
			t.Errorf("State(%d).String() = %q, want %q", state, got, want)
		}
	}
}
