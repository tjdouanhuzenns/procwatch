package supervisor

import (
	"sort"
	"testing"
)

func TestTagIndex_SetAndLookup(t *testing.T) {
	idx := NewTagIndex()
	_ = idx.Set("web", []string{"http", "frontend"})
	_ = idx.Set("api", []string{"http", "backend"})

	got := idx.Lookup("http")
	sort.Strings(got)
	if len(got) != 2 || got[0] != "api" || got[1] != "web" {
		t.Fatalf("expected [api web], got %v", got)
	}
}

func TestTagIndex_LookupMissingTag(t *testing.T) {
	idx := NewTagIndex()
	if got := idx.Lookup("nonexistent"); len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

func TestTagIndex_SetEmptyNameErrors(t *testing.T) {
	idx := NewTagIndex()
	if err := idx.Set("", []string{"tag"}); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestTagIndex_Remove(t *testing.T) {
	idx := NewTagIndex()
	_ = idx.Set("worker", []string{"background"})
	idx.Remove("worker")

	if got := idx.Lookup("background"); len(got) != 0 {
		t.Fatalf("expected empty after remove, got %v", got)
	}
	if got := idx.Tags("worker"); len(got) != 0 {
		t.Fatalf("expected no tags after remove, got %v", got)
	}
}

func TestTagIndex_SetReplacesOldTags(t *testing.T) {
	idx := NewTagIndex()
	_ = idx.Set("svc", []string{"old"})
	_ = idx.Set("svc", []string{"new"})

	if got := idx.Lookup("old"); len(got) != 0 {
		t.Fatalf("old tag should be gone, got %v", got)
	}
	if got := idx.Lookup("new"); len(got) != 1 || got[0] != "svc" {
		t.Fatalf("expected [svc] under new tag, got %v", got)
	}
}

func TestTagIndex_Tags(t *testing.T) {
	idx := NewTagIndex()
	_ = idx.Set("proc", []string{"a", "b", "c"})
	tags := idx.Tags("proc")
	sort.Strings(tags)
	if len(tags) != 3 || tags[0] != "a" || tags[1] != "b" || tags[2] != "c" {
		t.Fatalf("expected [a b c], got %v", tags)
	}
}

func TestTagIndex_AllTags(t *testing.T) {
	idx := NewTagIndex()
	_ = idx.Set("p1", []string{"x", "y"})
	_ = idx.Set("p2", []string{"y", "z"})
	all := idx.AllTags()
	sort.Strings(all)
	if len(all) != 3 || all[0] != "x" || all[1] != "y" || all[2] != "z" {
		t.Fatalf("expected [x y z], got %v", all)
	}
}

func TestTagIndex_SetEmptyTagsRemovesProcess(t *testing.T) {
	idx := NewTagIndex()
	_ = idx.Set("svc", []string{"env:prod"})
	_ = idx.Set("svc", []string{})

	if got := idx.Lookup("env:prod"); len(got) != 0 {
		t.Fatalf("expected no processes, got %v", got)
	}
}
