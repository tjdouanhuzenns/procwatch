package supervisor

import (
	"testing"
)

func TestProcessResource_SetAndGet(t *testing.T) {
	s := NewProcessResourceStore()
	err := s.Set("worker", ResourceLimits{MaxCPUPercent: 50.0, MaxMemoryMB: 256})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	l, ok := s.Get("worker")
	if !ok {
		t.Fatal("expected limits to be present")
	}
	if l.MaxCPUPercent != 50.0 {
		t.Errorf("expected MaxCPUPercent 50.0, got %v", l.MaxCPUPercent)
	}
	if l.MaxMemoryMB != 256 {
		t.Errorf("expected MaxMemoryMB 256, got %v", l.MaxMemoryMB)
	}
}

func TestProcessResource_GetMissing(t *testing.T) {
	s := NewProcessResourceStore()
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected missing entry")
	}
}

func TestProcessResource_SetEmptyNameErrors(t *testing.T) {
	s := NewProcessResourceStore()
	err := s.Set("", ResourceLimits{MaxMemoryMB: 128})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestProcessResource_SetBothZeroErrors(t *testing.T) {
	s := NewProcessResourceStore()
	err := s.Set("worker", ResourceLimits{})
	if err == nil {
		t.Fatal("expected error when both limits are zero")
	}
}

func TestProcessResource_SetNegativeCPUErrors(t *testing.T) {
	s := NewProcessResourceStore()
	err := s.Set("worker", ResourceLimits{MaxCPUPercent: -1.0, MaxMemoryMB: 128})
	if err == nil {
		t.Fatal("expected error for negative CPU percent")
	}
}

func TestProcessResource_Remove(t *testing.T) {
	s := NewProcessResourceStore()
	_ = s.Set("worker", ResourceLimits{MaxMemoryMB: 128})
	s.Remove("worker")
	_, ok := s.Get("worker")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestProcessResource_All(t *testing.T) {
	s := NewProcessResourceStore()
	_ = s.Set("a", ResourceLimits{MaxMemoryMB: 64})
	_ = s.Set("b", ResourceLimits{MaxCPUPercent: 25.0, MaxMemoryMB: 128})
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestProcessResource_TimestampIsSet(t *testing.T) {
	s := NewProcessResourceStore()
	_ = s.Set("worker", ResourceLimits{MaxMemoryMB: 256})
	l, _ := s.Get("worker")
	if l.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be set")
	}
}

func TestProcessResource_OverwriteUpdates(t *testing.T) {
	s := NewProcessResourceStore()
	_ = s.Set("worker", ResourceLimits{MaxMemoryMB: 64})
	_ = s.Set("worker", ResourceLimits{MaxMemoryMB: 512})
	l, _ := s.Get("worker")
	if l.MaxMemoryMB != 512 {
		t.Errorf("expected MaxMemoryMB 512, got %v", l.MaxMemoryMB)
	}
}
