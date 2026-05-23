package supervisor

import (
	"testing"
)

func TestProcessEnv_SetAndGet(t *testing.T) {
	s := NewProcessEnvStore()
	env := map[string]string{"FOO": "bar", "BAZ": "qux"}
	if err := s.Set("web", env); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := s.Get("web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["FOO"] != "bar" || got["BAZ"] != "qux" {
		t.Errorf("unexpected env: %v", got)
	}
}

func TestProcessEnv_GetMissing(t *testing.T) {
	s := NewProcessEnvStore()
	_, err := s.Get("missing")
	if err == nil {
		t.Fatal("expected error for missing process")
	}
}

func TestProcessEnv_SetEmptyNameErrors(t *testing.T) {
	s := NewProcessEnvStore()
	if err := s.Set("", map[string]string{"K": "V"}); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestProcessEnv_SetEmptyEnvErrors(t *testing.T) {
	s := NewProcessEnvStore()
	if err := s.Set("web", map[string]string{}); err == nil {
		t.Fatal("expected error for empty env map")
	}
}

func TestProcessEnv_SetEmptyKeyErrors(t *testing.T) {
	s := NewProcessEnvStore()
	if err := s.Set("web", map[string]string{"": "val"}); err == nil {
		t.Fatal("expected error for empty env key")
	}
}

func TestProcessEnv_Remove(t *testing.T) {
	s := NewProcessEnvStore()
	_ = s.Set("web", map[string]string{"A": "1"})
	s.Remove("web")
	if _, err := s.Get("web"); err == nil {
		t.Fatal("expected error after remove")
	}
}

func TestProcessEnv_All(t *testing.T) {
	s := NewProcessEnvStore()
	_ = s.Set("web", map[string]string{"A": "1"})
	_ = s.Set("worker", map[string]string{"B": "2"})
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}

func TestProcessEnv_SetOverwrites(t *testing.T) {
	s := NewProcessEnvStore()
	_ = s.Set("web", map[string]string{"A": "old"})
	_ = s.Set("web", map[string]string{"A": "new"})
	env, _ := s.Get("web")
	if env["A"] != "new" {
		t.Errorf("expected 'new', got %q", env["A"])
	}
}

func TestProcessEnv_IsolatedCopy(t *testing.T) {
	s := NewProcessEnvStore()
	original := map[string]string{"X": "1"}
	_ = s.Set("web", original)
	original["X"] = "mutated"
	env, _ := s.Get("web")
	if env["X"] != "1" {
		t.Errorf("store should not be affected by external mutation")
	}
}
