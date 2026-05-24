package supervisor

import (
	"testing"
	"time"
)

func TestProcessHeartbeat_BeatAndIsAlive(t *testing.T) {
	h := NewProcessHeartbeat()
	if err := h.Beat("svc", time.Second); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !h.IsAlive("svc") {
		t.Error("expected process to be alive immediately after beat")
	}
}

func TestProcessHeartbeat_StaleAfterThreshold(t *testing.T) {
	h := NewProcessHeartbeat()
	_ = h.Beat("svc", time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	if h.IsAlive("svc") {
		t.Error("expected process to be stale after threshold exceeded")
	}
}

func TestProcessHeartbeat_UnknownProcessNotAlive(t *testing.T) {
	h := NewProcessHeartbeat()
	if h.IsAlive("ghost") {
		t.Error("unknown process should not be alive")
	}
}

func TestProcessHeartbeat_EmptyNameErrors(t *testing.T) {
	h := NewProcessHeartbeat()
	if err := h.Beat("", time.Second); err == nil {
		t.Error("expected error for empty process name")
	}
}

func TestProcessHeartbeat_DefaultThreshold(t *testing.T) {
	h := NewProcessHeartbeat()
	_ = h.Beat("svc", 0)
	e, ok := h.Get("svc")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Threshold != 30*time.Second {
		t.Errorf("expected default threshold 30s, got %v", e.Threshold)
	}
}

func TestProcessHeartbeat_Remove(t *testing.T) {
	h := NewProcessHeartbeat()
	_ = h.Beat("svc", time.Second)
	h.Remove("svc")
	if _, ok := h.Get("svc"); ok {
		t.Error("expected entry to be removed")
	}
}

func TestProcessHeartbeat_All(t *testing.T) {
	h := NewProcessHeartbeat()
	_ = h.Beat("a", time.Second)
	_ = h.Beat("b", time.Second)
	all := h.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}

func TestProcessHeartbeat_BeatUpdatesTimestamp(t *testing.T) {
	h := NewProcessHeartbeat()
	_ = h.Beat("svc", time.Second)
	e1, _ := h.Get("svc")
	time.Sleep(2 * time.Millisecond)
	_ = h.Beat("svc", time.Second)
	e2, _ := h.Get("svc")
	if !e2.LastBeat.After(e1.LastBeat) {
		t.Error("expected updated timestamp after second beat")
	}
}
