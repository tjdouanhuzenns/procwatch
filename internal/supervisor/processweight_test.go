package supervisor

import (
	"testing"
)

func TestProcessWeight_SetAndGet(t *testing.T) {
	pw := NewProcessWeight()
	if err := pw.Set("worker", 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w, ok := pw.Get("worker")
	if !ok {
		t.Fatal("expected weight to be present")
	}
	if w != 10 {
		t.Fatalf("expected 10, got %d", w)
	}
}

func TestProcessWeight_GetMissing(t *testing.T) {
	pw := NewProcessWeight()
	_, ok := pw.Get("ghost")
	if ok {
		t.Fatal("expected missing entry")
	}
}

func TestProcessWeight_SetEmptyNameErrors(t *testing.T) {
	pw := NewProcessWeight()
	if err := pw.Set("", 5); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestProcessWeight_SetNegativeWeightErrors(t *testing.T) {
	pw := NewProcessWeight()
	if err := pw.Set("worker", -1); err == nil {
		t.Fatal("expected error for negative weight")
	}
}

func TestProcessWeight_ZeroWeightAllowed(t *testing.T) {
	pw := NewProcessWeight()
	if err := pw.Set("worker", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w, ok := pw.Get("worker")
	if !ok || w != 0 {
		t.Fatalf("expected 0, got %d ok=%v", w, ok)
	}
}

func TestProcessWeight_Remove(t *testing.T) {
	pw := NewProcessWeight()
	_ = pw.Set("worker", 5)
	pw.Remove("worker")
	_, ok := pw.Get("worker")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestProcessWeight_OverwriteUpdates(t *testing.T) {
	pw := NewProcessWeight()
	_ = pw.Set("worker", 3)
	_ = pw.Set("worker", 99)
	w, _ := pw.Get("worker")
	if w != 99 {
		t.Fatalf("expected 99, got %d", w)
	}
}

func TestProcessWeight_All(t *testing.T) {
	pw := NewProcessWeight()
	_ = pw.Set("a", 1)
	_ = pw.Set("b", 2)
	all := pw.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all["a"] != 1 || all["b"] != 2 {
		t.Fatalf("unexpected values: %v", all)
	}
}
