package supervisor

import (
	"testing"
)

func TestEnvOverride_SetAndGet(t *testing.T) {
	e := NewEnvOverride()
	if err := e.Set("svc", map[string]string{"LOG_LEVEL": "debug"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := e.Get("svc")
	if got["LOG_LEVEL"] != "debug" {
		t.Errorf("expected debug, got %q", got["LOG_LEVEL"])
	}
}

func TestEnvOverride_GetMissing(t *testing.T) {
	e := NewEnvOverride()
	got := e.Get("nonexistent")
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestEnvOverride_SetEmptyNameErrors(t *testing.T) {
	e := NewEnvOverride()
	if err := e.Set("", map[string]string{"K": "V"}); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestEnvOverride_SetEmptyVarsErrors(t *testing.T) {
	e := NewEnvOverride()
	if err := e.Set("svc", map[string]string{}); err == nil {
		t.Error("expected error for empty vars")
	}
}

func TestEnvOverride_SetEmptyKeyErrors(t *testing.T) {
	e := NewEnvOverride()
	if err := e.Set("svc", map[string]string{"": "value"}); err == nil {
		t.Error("expected error for empty key")
	}
}

func TestEnvOverride_SetMergesKeys(t *testing.T) {
	e := NewEnvOverride()
	_ = e.Set("svc", map[string]string{"A": "1"})
	_ = e.Set("svc", map[string]string{"B": "2"})
	got := e.Get("svc")
	if got["A"] != "1" || got["B"] != "2" {
		t.Errorf("expected merged keys, got %v", got)
	}
}

func TestEnvOverride_Remove(t *testing.T) {
	e := NewEnvOverride()
	_ = e.Set("svc", map[string]string{"K": "V"})
	e.Remove("svc")
	if len(e.Get("svc")) != 0 {
		t.Error("expected empty map after remove")
	}
}

func TestEnvOverride_All(t *testing.T) {
	e := NewEnvOverride()
	_ = e.Set("a", map[string]string{"X": "1"})
	_ = e.Set("b", map[string]string{"Y": "2"})
	all := e.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all["a"]["X"] != "1" || all["b"]["Y"] != "2" {
		t.Errorf("unexpected values in All: %v", all)
	}
}

func TestEnvOverride_AllReturnsCopy(t *testing.T) {
	e := NewEnvOverride()
	_ = e.Set("svc", map[string]string{"K": "original"})
	all := e.All()
	all["svc"]["K"] = "mutated"
	if e.Get("svc")["K"] != "original" {
		t.Error("All() should return a deep copy")
	}
}
