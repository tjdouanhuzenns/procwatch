package supervisor

import (
	"sync"
	"testing"
	"time"
)

func TestRetryPolicy_ConcurrentRecordAndAllow(t *testing.T) {
	m := NewRetryPolicyManager()
	_ = m.SetPolicy("svc", RetryPolicy{
		MaxAttempts: 100,
		Window:      time.Minute,
		Cooldown:    time.Second,
	})

	var wg sync.WaitGroup
	now := time.Now()

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Record("svc", now)
		}()
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Allow("svc", now)
		}()
	}

	wg.Wait()
}

func TestRetryPolicy_ConcurrentReset(t *testing.T) {
	m := NewRetryPolicyManager()
	_ = m.SetPolicy("svc", DefaultRetryPolicy())

	var wg sync.WaitGroup
	now := time.Now()

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Record("svc", now)
			m.Reset("svc")
		}()
	}

	wg.Wait()
	// After all resets, Allow should succeed
	if !m.Allow("svc", time.Now()) {
		t.Fatal("expected Allow after concurrent resets")
	}
}
