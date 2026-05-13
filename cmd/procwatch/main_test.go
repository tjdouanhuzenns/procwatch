package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMain_MissingConfig(t *testing.T) {
	if os.Getenv("RUN_MAIN") == "1" {
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMain_MissingConfig")
	cmd.Env = append(os.Environ(), "RUN_MAIN=1")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit for missing config")
	}
}

func TestMain_ValidConfig(t *testing.T) {
	if os.Getenv("RUN_MAIN") == "1" {
		main()
		return
	}

	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "procwatch.yaml")
	content := `processes:
  - name: greet
    command: echo
    args: ["hi"]
    restart: never
`
	if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_ValidConfig",
		"-config", cfgFile)
	cmd.Env = append(os.Environ(), "RUN_MAIN=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("output: %s", out)
		t.Fatalf("unexpected error: %v", err)
	}
}
