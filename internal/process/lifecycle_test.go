package process

import (
	"testing"
)

func TestLifecycleTracker_InitialState(t *testing.T) {
	lt := NewLifecycleTracker("svc")
	if lt.Current() != StateIdle {
		t.Fatalf("expected idle, got %s", lt.Current())
	}
}

func TestLifecycleTracker_ValidTransitions(t *testing.T) {
	lt := NewLifecycleTracker("svc")

	steps := []struct {
		to     State
		reason string
	}{
		{StateStarting, "launch"},
		{StateRunning, "pid acquired"},
		{StateStopping, "signal received"},
		{StateStopped, "exited 0"},
	}

	for _, step := range steps {
		if err := lt.Transition(step.to, step.reason); err != nil {
			t.Fatalf("unexpected error transitioning to %s: %v", step.to, err)
		}
	}

	if lt.Current() != StateStopped {
		t.Fatalf("expected stopped, got %s", lt.Current())
	}
}

func TestLifecycleTracker_InvalidTransition(t *testing.T) {
	lt := NewLifecycleTracker("svc")
	// Cannot jump directly from idle to running
	err := lt.Transition(StateRunning, "bad")
	if err == nil {
		t.Fatal("expected error for invalid transition, got nil")
	}
	// State must remain unchanged
	if lt.Current() != StateIdle {
		t.Fatalf("state should remain idle after failed transition, got %s", lt.Current())
	}
}

func TestLifecycleTracker_HistoryRecorded(t *testing.T) {
	lt := NewLifecycleTracker("svc")
	_ = lt.Transition(StateStarting, "boot")
	_ = lt.Transition(StateRunning, "ready")

	h := lt.History()
	if len(h) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(h))
	}
	if h[0].From != StateIdle || h[0].To != StateStarting {
		t.Errorf("unexpected first entry: %+v", h[0])
	}
	if h[1].From != StateStarting || h[1].To != StateRunning {
		t.Errorf("unexpected second entry: %+v", h[1])
	}
}

func TestLifecycleTracker_HistoryIsCopy(t *testing.T) {
	lt := NewLifecycleTracker("svc")
	_ = lt.Transition(StateStarting, "boot")
	h := lt.History()
	h[0].Reason = "mutated"

	h2 := lt.History()
	if h2[0].Reason == "mutated" {
		t.Error("History() should return a copy, not a reference")
	}
}

func TestStateString(t *testing.T) {
	cases := map[State]string{
		StateIdle:     "idle",
		StateStarting: "starting",
		StateRunning:  "running",
		StateStopping: "stopping",
		StateStopped:  "stopped",
		StateFailed:   "failed",
		StateBackoff:  "backoff",
	}
	for s, want := range cases {
		if got := s.String(); got != want {
			t.Errorf("State(%d).String() = %q, want %q", s, got, want)
		}
	}
}

func TestIsValidTransition_FailedCanRestart(t *testing.T) {
	if !isValidTransition(StateFailed, StateStarting) {
		t.Error("failed -> starting should be valid for restart policy")
	}
	if !isValidTransition(StateFailed, StateBackoff) {
		t.Error("failed -> backoff should be valid")
	}
}
