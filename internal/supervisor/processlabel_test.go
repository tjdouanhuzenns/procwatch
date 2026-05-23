package supervisor

import (
	"testing"
)

func TestProcessLabel_SetAndGet(t *testing.T) {
	s := NewProcessLabelStore()
	if err := s.Set("web", "env", "prod"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := s.Get("web", "env")
	if !ok || v != "prod" {
		t.Fatalf("expected prod, got %q ok=%v", v, ok)
	}
}

func TestProcessLabel_GetMissing(t *testing.T) {
	s := NewProcessLabelStore()
	_, ok := s.Get("ghost", "env")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestProcessLabel_SetEmptyNameErrors(t *testing.T) {
	s := NewProcessLabelStore()
	if err := s.Set("", "k", "v"); err == nil {
		t.Fatal("expected error for empty process name")
	}
}

func TestProcessLabel_SetEmptyKeyErrors(t *testing.T) {
	s := NewProcessLabelStore()
	if err := s.Set("web", "", "v"); err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestProcessLabel_All(t *testing.T) {
	s := NewProcessLabelStore()
	_ = s.Set("api", "region", "us-east")
	_ = s.Set("api", "tier", "backend")
	all := s.All("api")
	if len(all) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(all))
	}
	if all["region"] != "us-east" || all["tier"] != "backend" {
		t.Fatalf("unexpected labels: %v", all)
	}
}

func TestProcessLabel_Remove(t *testing.T) {
	s := NewProcessLabelStore()
	_ = s.Set("svc", "color", "blue")
	s.Remove("svc", "color")
	_, ok := s.Get("svc", "color")
	if ok {
		t.Fatal("expected label to be removed")
	}
}

func TestProcessLabel_RemoveAll(t *testing.T) {
	s := NewProcessLabelStore()
	_ = s.Set("svc", "a", "1")
	_ = s.Set("svc", "b", "2")
	s.RemoveAll("svc")
	if len(s.All("svc")) != 0 {
		t.Fatal("expected all labels removed")
	}
}

func TestProcessLabel_OverwriteUpdates(t *testing.T) {
	s := NewProcessLabelStore()
	_ = s.Set("worker", "stage", "v1")
	_ = s.Set("worker", "stage", "v2")
	v, _ := s.Get("worker", "stage")
	if v != "v2" {
		t.Fatalf("expected v2, got %q", v)
	}
}
