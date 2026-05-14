package supervisor

import (
	"testing"
	"time"
)

func TestThrottle_NotThrottledInitially(t *testing.T) {
	th := NewThrottle(DefaultThrottleConfig())
	if th.IsThrottled("svc") {
		t.Fatal("expected not throttled on fresh tracker")
	}
}

func TestThrottle_ThrottledAfterMaxStarts(t *testing.T) {
	cfg := ThrottleConfig{Window: 10 * time.Second, MaxStarts: 3}
	th := NewThrottle(cfg)

	for i := 0; i < 3; i++ {
		th.Record("svc")
	}

	if !th.IsThrottled("svc") {
		t.Fatal("expected throttled after max starts")
	}
}

func TestThrottle_NotThrottledBelowMax(t *testing.T) {
	cfg := ThrottleConfig{Window: 10 * time.Second, MaxStarts: 3}
	th := NewThrottle(cfg)

	th.Record("svc")
	th.Record("svc")

	if th.IsThrottled("svc") {
		t.Fatal("expected not throttled below max starts")
	}
}

func TestThrottle_OldEventsExpire(t *testing.T) {
	cfg := ThrottleConfig{Window: 5 * time.Second, MaxStarts: 3}
	th := NewThrottle(cfg)

	base := time.Now()
	// Simulate old events outside the window
	th.now = func() time.Time { return base.Add(-10 * time.Second) }
	th.Record("svc")
	th.Record("svc")
	th.Record("svc")

	// Advance time so old events are outside the window
	th.now = func() time.Time { return base }

	if th.IsThrottled("svc") {
		t.Fatal("expected old events to expire outside window")
	}
}

func TestThrottle_Reset(t *testing.T) {
	cfg := ThrottleConfig{Window: 10 * time.Second, MaxStarts: 2}
	th := NewThrottle(cfg)

	th.Record("svc")
	th.Record("svc")

	if !th.IsThrottled("svc") {
		t.Fatal("expected throttled before reset")
	}

	th.Reset("svc")

	if th.IsThrottled("svc") {
		t.Fatal("expected not throttled after reset")
	}
}

func TestThrottle_IndependentProcesses(t *testing.T) {
	cfg := ThrottleConfig{Window: 10 * time.Second, MaxStarts: 2}
	th := NewThrottle(cfg)

	th.Record("svcA")
	th.Record("svcA")

	if !th.IsThrottled("svcA") {
		t.Fatal("expected svcA to be throttled")
	}
	if th.IsThrottled("svcB") {
		t.Fatal("expected svcB to be unaffected")
	}
}
