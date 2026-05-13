package supervisor

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestAlerter_WritesOnMatchingState(t *testing.T) {
	reporter := NewStatusReporter()
	reporter.Update("svc", ProcessStatus{State: "running"})

	w := NewWatcher(reporter, 20*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)

	var buf bytes.Buffer
	alerter := NewAlerter(w, AlertRule{TriggerStates: []string{"running"}}, &buf)
	go alerter.Run(ctx)

	time.Sleep(100 * time.Millisecond)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatal("expected at least one alert line")
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["process"] != "svc" {
		t.Fatalf("expected process=svc, got %v", m["process"])
	}
	if m["alert"] != true {
		t.Fatal("expected alert=true")
	}
}

func TestAlerter_SilentOnNonMatchingState(t *testing.T) {
	reporter := NewStatusReporter()
	reporter.Update("svc", ProcessStatus{State: "running"})

	w := NewWatcher(reporter, 20*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)

	var buf bytes.Buffer
	// Only alert on "crashed", not "running".
	alerter := NewAlerter(w, AlertRule{TriggerStates: []string{"crashed"}}, &buf)
	go alerter.Run(ctx)

	time.Sleep(100 * time.Millisecond)

	if buf.Len() != 0 {
		t.Fatalf("expected no output, got: %s", buf.String())
	}
}

func TestAlerter_StopsOnContextCancel(t *testing.T) {
	reporter := NewStatusReporter()
	w := NewWatcher(reporter, 20*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())

	var buf bytes.Buffer
	alerter := NewAlerter(w, AlertRule{TriggerStates: []string{"running"}}, &buf)

	done := make(chan struct{})
	go func() {
		alerter.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("alerter did not stop after context cancel")
	}
}
