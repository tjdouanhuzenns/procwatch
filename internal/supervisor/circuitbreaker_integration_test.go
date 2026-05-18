package supervisor

import (
	"sync"
	"testing"
	"time"
)

// TestCircuitBreaker_ConcurrentAccess ensures the circuit breaker is safe
// for concurrent use across multiple goroutines recording failures and
// checking Allow simultaneously.
func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 10,
		SuccessThreshold: 2,
		OpenDuration:     50 * time.Millisecond,
	})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cb.RecordFailure("svc")
			cb.Allow("svc")
		}()
	}
	wg.Wait()

	// After 20 failures against threshold of 10, circuit must be open.
	if cb.State("svc") != CircuitOpen {
		t.Fatalf("expected open after concurrent failures, got %s", cb.State("svc"))
	}
}

// TestCircuitBreaker_ReopenOnHalfOpenFailure verifies that a failure during
// the half-open probe re-opens the circuit immediately.
func TestCircuitBreaker_ReopenOnHalfOpenFailure(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 3,
		OpenDuration:     50 * time.Millisecond,
	})

	cb.RecordFailure("svc")
	cb.RecordFailure("svc")
	if cb.State("svc") != CircuitOpen {
		t.Fatal("expected open")
	}

	time.Sleep(60 * time.Millisecond)
	if !cb.Allow("svc") {
		t.Fatal("expected allow in half-open")
	}
	if cb.State("svc") != CircuitHalfOpen {
		t.Fatalf("expected half-open, got %s", cb.State("svc"))
	}

	// Failure during half-open should re-open.
	cb.RecordFailure("svc")
	if cb.State("svc") != CircuitOpen {
		t.Fatalf("expected re-opened, got %s", cb.State("svc"))
	}
	if cb.Allow("svc") {
		t.Fatal("expected blocked after re-open")
	}
}
