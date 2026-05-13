package supervisor

import (
	"testing"
	"time"
)

func TestStatusReporter_UpdateAndGet(t *testing.T) {
	r := NewStatusReporter()
	now := time.Now()

	r.Update("web", "running", 1234, 0, now, time.Time{})

	ps, ok := r.Get("web")
	if !ok {
		t.Fatal("expected status to exist")
	}
	if ps.Name != "web" {
		t.Errorf("expected name 'web', got %q", ps.Name)
	}
	if ps.State != "running" {
		t.Errorf("expected state 'running', got %q", ps.State)
	}
	if ps.PID != 1234 {
		t.Errorf("expected PID 1234, got %d", ps.PID)
	}
	if ps.Restarts != 0 {
		t.Errorf("expected 0 restarts, got %d", ps.Restarts)
	}
}

func TestStatusReporter_GetMissing(t *testing.T) {
	r := NewStatusReporter()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("expected Get to return false for missing process")
	}
}

func TestStatusReporter_All(t *testing.T) {
	r := NewStatusReporter()
	now := time.Now()
	r.Update("svc-a", "running", 100, 1, now, time.Time{})
	r.Update("svc-b", "stopped", 0, 3, now, now)

	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(all))
	}
}

func TestStatusReporter_Remove(t *testing.T) {
	r := NewStatusReporter()
	now := time.Now()
	r.Update("ephemeral", "stopped", 0, 0, now, now)
	r.Remove("ephemeral")

	_, ok := r.Get("ephemeral")
	if ok {
		t.Error("expected status to be removed")
	}
}

func TestStatusReporter_UpdateOverwrites(t *testing.T) {
	r := NewStatusReporter()
	now := time.Now()
	r.Update("worker", "starting", 0, 0, now, time.Time{})
	r.Update("worker", "running", 5678, 1, now, time.Time{})

	ps, ok := r.Get("worker")
	if !ok {
		t.Fatal("expected status to exist")
	}
	if ps.State != "running" {
		t.Errorf("expected state 'running', got %q", ps.State)
	}
	if ps.PID != 5678 {
		t.Errorf("expected PID 5678, got %d", ps.PID)
	}
	if ps.Restarts != 1 {
		t.Errorf("expected 1 restart, got %d", ps.Restarts)
	}
}
