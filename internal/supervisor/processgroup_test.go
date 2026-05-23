package supervisor

import (
	"sort"
	"testing"
)

func TestProcessGroup_AddAndMembers(t *testing.T) {
	pg := NewProcessGroup()
	if err := pg.Add("web", "nginx"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := pg.Add("web", "caddy"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	members, err := pg.Members("web")
	if err != nil {
		t.Fatalf("Members: %v", err)
	}
	sort.Strings(members)
	if len(members) != 2 || members[0] != "caddy" || members[1] != "nginx" {
		t.Errorf("unexpected members: %v", members)
	}
}

func TestProcessGroup_GetMissingGroup(t *testing.T) {
	pg := NewProcessGroup()
	_, err := pg.Members("missing")
	if err == nil {
		t.Fatal("expected error for missing group")
	}
}

func TestProcessGroup_AddEmptyGroupErrors(t *testing.T) {
	pg := NewProcessGroup()
	if err := pg.Add("", "nginx"); err == nil {
		t.Fatal("expected error for empty group")
	}
}

func TestProcessGroup_AddEmptyProcessErrors(t *testing.T) {
	pg := NewProcessGroup()
	if err := pg.Add("web", ""); err == nil {
		t.Fatal("expected error for empty process")
	}
}

func TestProcessGroup_Remove(t *testing.T) {
	pg := NewProcessGroup()
	_ = pg.Add("web", "nginx")
	_ = pg.Add("web", "caddy")
	if err := pg.Remove("web", "nginx"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	members, _ := pg.Members("web")
	if len(members) != 1 || members[0] != "caddy" {
		t.Errorf("unexpected members after remove: %v", members)
	}
}

func TestProcessGroup_RemoveLastDeletesGroup(t *testing.T) {
	pg := NewProcessGroup()
	_ = pg.Add("web", "nginx")
	_ = pg.Remove("web", "nginx")
	groups := pg.Groups()
	if len(groups) != 0 {
		t.Errorf("expected group to be deleted, got: %v", groups)
	}
}

func TestProcessGroup_RemoveMissingGroupErrors(t *testing.T) {
	pg := NewProcessGroup()
	if err := pg.Remove("missing", "nginx"); err == nil {
		t.Fatal("expected error for missing group")
	}
}

func TestProcessGroup_Groups(t *testing.T) {
	pg := NewProcessGroup()
	_ = pg.Add("web", "nginx")
	_ = pg.Add("db", "postgres")
	groups := pg.Groups()
	sort.Strings(groups)
	if len(groups) != 2 || groups[0] != "db" || groups[1] != "web" {
		t.Errorf("unexpected groups: %v", groups)
	}
}

func TestProcessGroup_DeleteGroup(t *testing.T) {
	pg := NewProcessGroup()
	_ = pg.Add("web", "nginx")
	_ = pg.Add("web", "caddy")
	if err := pg.DeleteGroup("web"); err != nil {
		t.Fatalf("DeleteGroup: %v", err)
	}
	if len(pg.Groups()) != 0 {
		t.Error("expected no groups after DeleteGroup")
	}
}

func TestProcessGroup_DeleteMissingGroupErrors(t *testing.T) {
	pg := NewProcessGroup()
	if err := pg.DeleteGroup("missing"); err == nil {
		t.Fatal("expected error deleting missing group")
	}
}
