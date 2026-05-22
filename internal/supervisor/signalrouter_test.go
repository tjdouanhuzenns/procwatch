package supervisor

import (
	"testing"
)

func TestSignalRouter_RegisterAndLookup(t *testing.T) {
	r := NewSignalRouter()
	if err := r.Register("SIGUSR1", "worker", "reload"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	actions := r.Lookup("SIGUSR1")
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].ProcessName != "worker" || actions[0].Signal != "reload" {
		t.Errorf("unexpected action: %+v", actions[0])
	}
}

func TestSignalRouter_LookupMissingSignal(t *testing.T) {
	r := NewSignalRouter()
	actions := r.Lookup("SIGTERM")
	if actions != nil {
		t.Errorf("expected nil for unknown signal, got %v", actions)
	}
}

func TestSignalRouter_RegisterEmptySignalErrors(t *testing.T) {
	r := NewSignalRouter()
	if err := r.Register("", "worker", "reload"); err == nil {
		t.Error("expected error for empty signal")
	}
}

func TestSignalRouter_RegisterEmptyProcessErrors(t *testing.T) {
	r := NewSignalRouter()
	if err := r.Register("SIGUSR1", "", "reload"); err == nil {
		t.Error("expected error for empty process name")
	}
}

func TestSignalRouter_RegisterEmptyActionErrors(t *testing.T) {
	r := NewSignalRouter()
	if err := r.Register("SIGUSR1", "worker", ""); err == nil {
		t.Error("expected error for empty action")
	}
}

func TestSignalRouter_MultipleActionsPerSignal(t *testing.T) {
	r := NewSignalRouter()
	_ = r.Register("SIGUSR1", "worker", "reload")
	_ = r.Register("SIGUSR1", "cache", "flush")
	actions := r.Lookup("SIGUSR1")
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
}

func TestSignalRouter_Remove(t *testing.T) {
	r := NewSignalRouter()
	_ = r.Register("SIGUSR1", "worker", "reload")
	r.Remove("SIGUSR1")
	if actions := r.Lookup("SIGUSR1"); actions != nil {
		t.Errorf("expected nil after remove, got %v", actions)
	}
}

func TestSignalRouter_All(t *testing.T) {
	r := NewSignalRouter()
	_ = r.Register("SIGUSR1", "worker", "reload")
	_ = r.Register("SIGTERM", "db", "shutdown")
	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 signals, got %d", len(all))
	}
	if _, ok := all["SIGUSR1"]; !ok {
		t.Error("expected SIGUSR1 in all")
	}
}

func TestSignalRouter_LookupReturnsCopy(t *testing.T) {
	r := NewSignalRouter()
	_ = r.Register("SIGUSR1", "worker", "reload")
	actions := r.Lookup("SIGUSR1")
	actions[0].ProcessName = "mutated"
	original := r.Lookup("SIGUSR1")
	if original[0].ProcessName == "mutated" {
		t.Error("Lookup should return a copy, not a reference")
	}
}
