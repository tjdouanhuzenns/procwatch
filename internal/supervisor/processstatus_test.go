package supervisor

import (
	"testing"
	"time"
)

func TestProcessStatus_SetAndGet(t *testing.T) {
	s := NewProcessStatusStore()
	if err := s.Set("svc", "running"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := s.Get("svc")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Status != "running" {
		t.Errorf("expected 'running', got %q", e.Status)
	}
}

func TestProcessStatus_GetMissing(t *testing.T) {
	s := NewProcessStatusStore()
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected missing entry")
	}
}

func TestProcessStatus_SetEmptyNameErrors(t *testing.T) {
	s := NewProcessStatusStore()
	if err := s.Set("", "running"); err == nil {
		t.Fatal("expected error for empty process name")
	}
}

func TestProcessStatus_SetEmptyStatusErrors(t *testing.T) {
	s := NewProcessStatusStore()
	if err := s.Set("svc", ""); err == nil {
		t.Fatal("expected error for empty status")
	}
}

func TestProcessStatus_TimestampIsSet(t *testing.T) {
	before := time.Now()
	s := NewProcessStatusStore()
	_ = s.Set("svc", "starting")
	e, _ := s.Get("svc")
	if e.UpdatedAt.Before(before) {
		t.Errorf("timestamp %v is before test start %v", e.UpdatedAt, before)
	}
}

func TestProcessStatus_OverwriteUpdates(t *testing.T) {
	s := NewProcessStatusStore()
	_ = s.Set("svc", "starting")
	_ = s.Set("svc", "running")
	e, _ := s.Get("svc")
	if e.Status != "running" {
		t.Errorf("expected 'running', got %q", e.Status)
	}
}

func TestProcessStatus_Remove(t *testing.T) {
	s := NewProcessStatusStore()
	_ = s.Set("svc", "running")
	s.Remove("svc")
	_, ok := s.Get("svc")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestProcessStatus_All(t *testing.T) {
	s := NewProcessStatusStore()
	_ = s.Set("a", "running")
	_ = s.Set("b", "stopped")
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}
