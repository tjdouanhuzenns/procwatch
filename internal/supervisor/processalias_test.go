package supervisor

import (
	"testing"
)

func TestProcessAlias_SetAndResolve(t *testing.T) {
	pa := NewProcessAlias()
	if err := pa.Set("web", "nginx"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	name, err := pa.Resolve("web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "nginx" {
		t.Errorf("expected nginx, got %q", name)
	}
}

func TestProcessAlias_ResolveMissing(t *testing.T) {
	pa := NewProcessAlias()
	_, err := pa.Resolve("ghost")
	if err == nil {
		t.Error("expected error for missing alias")
	}
}

func TestProcessAlias_SetEmptyAliasErrors(t *testing.T) {
	pa := NewProcessAlias()
	if err := pa.Set("", "nginx"); err == nil {
		t.Error("expected error for empty alias")
	}
}

func TestProcessAlias_SetEmptyNameErrors(t *testing.T) {
	pa := NewProcessAlias()
	if err := pa.Set("web", ""); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestProcessAlias_DuplicateAliasConflictErrors(t *testing.T) {
	pa := NewProcessAlias()
	_ = pa.Set("web", "nginx")
	if err := pa.Set("web", "apache"); err == nil {
		t.Error("expected error when alias maps to different process")
	}
}

func TestProcessAlias_DuplicateAliasIdempotent(t *testing.T) {
	pa := NewProcessAlias()
	_ = pa.Set("web", "nginx")
	if err := pa.Set("web", "nginx"); err != nil {
		t.Errorf("re-setting same alias should be idempotent, got: %v", err)
	}
}

func TestProcessAlias_AliasesFor(t *testing.T) {
	pa := NewProcessAlias()
	_ = pa.Set("web", "nginx")
	_ = pa.Set("frontend", "nginx")
	aliases := pa.AliasesFor("nginx")
	if len(aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(aliases))
	}
}

func TestProcessAlias_Remove(t *testing.T) {
	pa := NewProcessAlias()
	_ = pa.Set("web", "nginx")
	pa.Remove("web")
	_, err := pa.Resolve("web")
	if err == nil {
		t.Error("expected error after removal")
	}
	aliases := pa.AliasesFor("nginx")
	if len(aliases) != 0 {
		t.Errorf("expected 0 aliases after removal, got %d", len(aliases))
	}
}

func TestProcessAlias_RemoveUnknownIsNoop(t *testing.T) {
	pa := NewProcessAlias()
	pa.Remove("nonexistent") // should not panic
}

func TestProcessAlias_All(t *testing.T) {
	pa := NewProcessAlias()
	_ = pa.Set("web", "nginx")
	_ = pa.Set("db", "postgres")
	all := pa.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
	if all["web"] != "nginx" {
		t.Errorf("expected nginx for web, got %q", all["web"])
	}
}
