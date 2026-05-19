package supervisor

import (
	"sort"
	"testing"
)

func TestLabelFilter_SetAndGet(t *testing.T) {
	lf := NewLabelFilter()
	err := lf.Set("worker", map[string]string{"env": "prod", "tier": "backend"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := lf.Get("worker")
	if got["env"] != "prod" || got["tier"] != "backend" {
		t.Errorf("unexpected labels: %v", got)
	}
}

func TestLabelFilter_GetMissing(t *testing.T) {
	lf := NewLabelFilter()
	if lf.Get("nonexistent") != nil {
		t.Error("expected nil for missing process")
	}
}

func TestLabelFilter_SetEmptyNameErrors(t *testing.T) {
	lf := NewLabelFilter()
	if err := lf.Set("", map[string]string{"k": "v"}); err == nil {
		t.Error("expected error for empty process name")
	}
}

func TestLabelFilter_Remove(t *testing.T) {
	lf := NewLabelFilter()
	_ = lf.Set("svc", map[string]string{"env": "staging"})
	lf.Remove("svc")
	if lf.Get("svc") != nil {
		t.Error("expected nil after remove")
	}
}

func TestLabelFilter_MatchAll(t *testing.T) {
	lf := NewLabelFilter()
	_ = lf.Set("a", map[string]string{"env": "prod"})
	_ = lf.Set("b", map[string]string{"env": "staging"})
	_ = lf.Set("c", map[string]string{"env": "prod", "tier": "frontend"})

	matches := lf.Match(map[string]string{"env": "prod"})
	sort.Strings(matches)
	if len(matches) != 2 || matches[0] != "a" || matches[1] != "c" {
		t.Errorf("unexpected matches: %v", matches)
	}
}

func TestLabelFilter_MatchEmptySelectorReturnsAll(t *testing.T) {
	lf := NewLabelFilter()
	_ = lf.Set("x", map[string]string{"k": "v"})
	_ = lf.Set("y", map[string]string{"k": "w"})

	matches := lf.Match(map[string]string{})
	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
}

func TestLabelFilter_GetReturnsCopy(t *testing.T) {
	lf := NewLabelFilter()
	_ = lf.Set("p", map[string]string{"key": "original"})
	got := lf.Get("p")
	got["key"] = "mutated"
	got2 := lf.Get("p")
	if got2["key"] != "original" {
		t.Error("Get should return a copy, not a reference")
	}
}
