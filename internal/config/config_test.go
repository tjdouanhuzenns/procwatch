package config

import (
	"os"
	"testing"
	"time"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "procwatch-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	path := writeTemp(t, `
log_format: json
processes:
  - name: worker
    command: /usr/bin/worker
    args: ["--verbose"]
    restart_policy: always
    max_restarts: 3
    backoff: 1s
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("expected log_format json, got %q", cfg.LogFormat)
	}
	if len(cfg.Processes) != 1 {
		t.Fatalf("expected 1 process, got %d", len(cfg.Processes))
	}
	p := cfg.Processes[0]
	if p.Name != "worker" {
		t.Errorf("expected name worker, got %q", p.Name)
	}
	if p.RestartPolicy != RestartAlways {
		t.Errorf("expected restart_policy always, got %q", p.RestartPolicy)
	}
	if p.Backoff != time.Second {
		t.Errorf("expected backoff 1s, got %v", p.Backoff)
	}
}

func TestLoad_Defaults(t *testing.T) {
	path := writeTemp(t, `
processes:
  - name: svc
    command: /bin/svc
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("default log_format should be json, got %q", cfg.LogFormat)
	}
	p := cfg.Processes[0]
	if p.RestartPolicy != RestartOnFailure {
		t.Errorf("default restart_policy should be on-failure, got %q", p.RestartPolicy)
	}
	if p.MaxRestarts != 5 {
		t.Errorf("default max_restarts should be 5, got %d", p.MaxRestarts)
	}
	if p.Backoff != 2*time.Second {
		t.Errorf("default backoff should be 2s, got %v", p.Backoff)
	}
}

func TestLoad_DuplicateName(t *testing.T) {
	path := writeTemp(t, `
processes:
  - name: dup
    command: /bin/a
  - name: dup
    command: /bin/b
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate process name")
	}
}

func TestLoad_MissingCommand(t *testing.T) {
	path := writeTemp(t, `
processes:
  - name: nocommand
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing command")
	}
}

func TestLoad_InvalidLogFormat(t *testing.T) {
	path := writeTemp(t, `
log_format: xml
processes:
  - name: svc
    command: /bin/svc
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid log_format")
	}
}
