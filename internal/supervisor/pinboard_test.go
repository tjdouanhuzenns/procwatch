package supervisor

import (
	"testing"
)

func TestPinboard_SetAndGet(t *testing.T) {
	pb := NewPinboard()
	if err := pb.Set("web", "owner", "platform-team"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := pb.Get("web", "owner")
	if !ok {
		t.Fatal("expected annotation to exist")
	}
	if v != "platform-team" {
		t.Errorf("expected 'platform-team', got %q", v)
	}
}

func TestPinboard_GetMissing(t *testing.T) {
	pb := NewPinboard()
	_, ok := pb.Get("ghost", "key")
	if ok {
		t.Error("expected missing annotation to return false")
	}
}

func TestPinboard_SetEmptyNameErrors(t *testing.T) {
	pb := NewPinboard()
	if err := pb.Set("", "k", "v"); err == nil {
		t.Error("expected error for empty process name")
	}
}

func TestPinboard_SetEmptyKeyErrors(t *testing.T) {
	pb := NewPinboard()
	if err := pb.Set("web", "", "v"); err == nil {
		t.Error("expected error for empty key")
	}
}

func TestPinboard_All(t *testing.T) {
	pb := NewPinboard()
	_ = pb.Set("svc", "a", "1")
	_ = pb.Set("svc", "b", "2")
	all := pb.All("svc")
	if len(all) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(all))
	}
	if all["a"] != "1" || all["b"] != "2" {
		t.Errorf("unexpected annotations: %v", all)
	}
}

func TestPinboard_AllMissing(t *testing.T) {
	pb := NewPinboard()
	if pb.All("nobody") != nil {
		t.Error("expected nil for missing process")
	}
}

func TestPinboard_AllIsCopy(t *testing.T) {
	pb := NewPinboard()
	_ = pb.Set("svc", "x", "original")
	all := pb.All("svc")
	all["x"] = "mutated"
	v, _ := pb.Get("svc", "x")
	if v != "original" {
		t.Error("All() should return a copy, not a reference")
	}
}

func TestPinboard_Delete(t *testing.T) {
	pb := NewPinboard()
	_ = pb.Set("svc", "k", "v")
	pb.Delete("svc", "k")
	_, ok := pb.Get("svc", "k")
	if ok {
		t.Error("expected annotation to be deleted")
	}
}

func TestPinboard_Remove(t *testing.T) {
	pb := NewPinboard()
	_ = pb.Set("svc", "k", "v")
	pb.Remove("svc")
	if pb.All("svc") != nil {
		t.Error("expected all annotations removed")
	}
}

func TestPinboard_OverwriteValue(t *testing.T) {
	pb := NewPinboard()
	_ = pb.Set("svc", "env", "staging")
	_ = pb.Set("svc", "env", "production")
	v, _ := pb.Get("svc", "env")
	if v != "production" {
		t.Errorf("expected 'production', got %q", v)
	}
}
