package supervisor

import (
	"testing"
)

func TestProcessMetadata_SetAndGet(t *testing.T) {
	s := NewProcessMetadataStore()
	if err := s.Set("web", "region", "us-east-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry, ok := s.Get("web", "region")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if entry.Value != "us-east-1" {
		t.Errorf("expected us-east-1, got %s", entry.Value)
	}
	if entry.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestProcessMetadata_GetMissing(t *testing.T) {
	s := NewProcessMetadataStore()
	_, ok := s.Get("missing", "key")
	if ok {
		t.Error("expected missing entry to not exist")
	}
}

func TestProcessMetadata_SetEmptyProcessErrors(t *testing.T) {
	s := NewProcessMetadataStore()
	if err := s.Set("", "key", "val"); err == nil {
		t.Error("expected error for empty process name")
	}
}

func TestProcessMetadata_SetEmptyKeyErrors(t *testing.T) {
	s := NewProcessMetadataStore()
	if err := s.Set("web", "", "val"); err == nil {
		t.Error("expected error for empty key")
	}
}

func TestProcessMetadata_OverwriteUpdates(t *testing.T) {
	s := NewProcessMetadataStore()
	_ = s.Set("web", "env", "staging")
	_ = s.Set("web", "env", "production")
	entry, ok := s.Get("web", "env")
	if !ok {
		t.Fatal("expected entry")
	}
	if entry.Value != "production" {
		t.Errorf("expected production, got %s", entry.Value)
	}
}

func TestProcessMetadata_ForProcess(t *testing.T) {
	s := NewProcessMetadataStore()
	_ = s.Set("web", "region", "us-east-1")
	_ = s.Set("web", "env", "prod")
	_ = s.Set("worker", "env", "prod")
	entries := s.ForProcess("web")
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for web, got %d", len(entries))
	}
}

func TestProcessMetadata_Delete(t *testing.T) {
	s := NewProcessMetadataStore()
	_ = s.Set("web", "region", "us-east-1")
	s.Delete("web", "region")
	_, ok := s.Get("web", "region")
	if ok {
		t.Error("expected entry to be deleted")
	}
}

func TestProcessMetadata_All(t *testing.T) {
	s := NewProcessMetadataStore()
	_ = s.Set("web", "region", "us-east-1")
	_ = s.Set("worker", "env", "prod")
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 total entries, got %d", len(all))
	}
}

func TestProcessMetadata_DeleteCleansUpProcess(t *testing.T) {
	s := NewProcessMetadataStore()
	_ = s.Set("web", "only", "val")
	s.Delete("web", "only")
	entries := s.ForProcess("web")
	if len(entries) != 0 {
		t.Errorf("expected no entries after last key deleted, got %d", len(entries))
	}
}
