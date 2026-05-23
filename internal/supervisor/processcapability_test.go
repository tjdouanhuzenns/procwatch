package supervisor

import (
	"testing"
)

func TestProcessCapability_SetAndGet(t *testing.T) {
	s := NewProcessCapabilityStore()
	if err := s.Set("web", []string{"net_bind", "read_logs"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	caps, ok := s.Get("web")
	if !ok {
		t.Fatal("expected capabilities to be found")
	}
	if len(caps) != 2 || caps[0] != "net_bind" || caps[1] != "read_logs" {
		t.Errorf("unexpected caps: %v", caps)
	}
}

func TestProcessCapability_GetMissing(t *testing.T) {
	s := NewProcessCapabilityStore()
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected missing process to return false")
	}
}

func TestProcessCapability_SetEmptyNameErrors(t *testing.T) {
	s := NewProcessCapabilityStore()
	if err := s.Set("", []string{"cap"}); err == nil {
		t.Fatal("expected error for empty process name")
	}
}

func TestProcessCapability_SetEmptyCapsErrors(t *testing.T) {
	s := NewProcessCapabilityStore()
	if err := s.Set("web", []string{}); err == nil {
		t.Fatal("expected error for empty capabilities slice")
	}
}

func TestProcessCapability_SetEmptyCapStringErrors(t *testing.T) {
	s := NewProcessCapabilityStore()
	if err := s.Set("web", []string{"net_bind", ""}); err == nil {
		t.Fatal("expected error for empty capability string")
	}
}

func TestProcessCapability_Has(t *testing.T) {
	s := NewProcessCapabilityStore()
	_ = s.Set("worker", []string{"write_db"})
	if !s.Has("worker", "write_db") {
		t.Error("expected Has to return true for assigned capability")
	}
	if s.Has("worker", "net_bind") {
		t.Error("expected Has to return false for unassigned capability")
	}
	if s.Has("ghost", "write_db") {
		t.Error("expected Has to return false for unknown process")
	}
}

func TestProcessCapability_Remove(t *testing.T) {
	s := NewProcessCapabilityStore()
	_ = s.Set("api", []string{"net_bind"})
	s.Remove("api")
	_, ok := s.Get("api")
	if ok {
		t.Fatal("expected capabilities to be removed")
	}
}

func TestProcessCapability_All(t *testing.T) {
	s := NewProcessCapabilityStore()
	_ = s.Set("a", []string{"x"})
	_ = s.Set("b", []string{"y", "z"})
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestProcessCapability_SetReplacesOld(t *testing.T) {
	s := NewProcessCapabilityStore()
	_ = s.Set("svc", []string{"old_cap"})
	_ = s.Set("svc", []string{"new_cap"})
	caps, _ := s.Get("svc")
	if len(caps) != 1 || caps[0] != "new_cap" {
		t.Errorf("expected capabilities to be replaced, got %v", caps)
	}
}
