package supervisor

import (
	"testing"
	"time"
)

func TestProcessVersion_SetAndGet(t *testing.T) {
	s := NewProcessVersionStore()
	built := time.Now().Add(-24 * time.Hour)
	if err := s.Set("api", "v1.2.3", "abc123", built); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := s.Get("api")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if v.Version != "v1.2.3" {
		t.Errorf("expected v1.2.3, got %s", v.Version)
	}
	if v.Commit != "abc123" {
		t.Errorf("expected abc123, got %s", v.Commit)
	}
	if !v.BuiltAt.Equal(built) {
		t.Errorf("built-at mismatch")
	}
}

func TestProcessVersion_GetMissing(t *testing.T) {
	s := NewProcessVersionStore()
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected miss")
	}
}

func TestProcessVersion_SetEmptyNameErrors(t *testing.T) {
	s := NewProcessVersionStore()
	if err := s.Set("", "v1.0.0", "", time.Now()); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestProcessVersion_SetEmptyVersionErrors(t *testing.T) {
	s := NewProcessVersionStore()
	if err := s.Set("api", "", "", time.Now()); err == nil {
		t.Fatal("expected error for empty version")
	}
}

func TestProcessVersion_UpdatedAtIsSet(t *testing.T) {
	s := NewProcessVersionStore()
	before := time.Now()
	_ = s.Set("api", "v1.0.0", "", time.Time{})
	v, _ := s.Get("api")
	if v.UpdatedAt.Before(before) {
		t.Errorf("UpdatedAt not set correctly")
	}
}

func TestProcessVersion_Remove(t *testing.T) {
	s := NewProcessVersionStore()
	_ = s.Set("api", "v1.0.0", "", time.Now())
	s.Remove("api")
	_, ok := s.Get("api")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestProcessVersion_All(t *testing.T) {
	s := NewProcessVersionStore()
	_ = s.Set("api", "v1.0.0", "", time.Now())
	_ = s.Set("worker", "v2.1.0", "deadbeef", time.Now())
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}

func TestProcessVersion_OverwriteUpdates(t *testing.T) {
	s := NewProcessVersionStore()
	_ = s.Set("api", "v1.0.0", "aaa", time.Now())
	_ = s.Set("api", "v2.0.0", "bbb", time.Now())
	v, _ := s.Get("api")
	if v.Version != "v2.0.0" {
		t.Errorf("expected v2.0.0 after overwrite, got %s", v.Version)
	}
}
