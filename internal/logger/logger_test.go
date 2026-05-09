package logger_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/example/procwatch/internal/logger"
)

func decode(t *testing.T, buf *bytes.Buffer) logger.Entry {
	t.Helper()
	var e logger.Entry
	if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
		t.Fatalf("failed to decode log entry: %v (raw: %s)", err, buf.String())
	}
	return e
}

func TestLogger_Levels(t *testing.T) {
	tests := []struct {
		name    string
		log     func(*logger.Logger, string)
		wantLvl logger.Level
	}{
		{"info", (*logger.Logger).Info, logger.LevelInfo},
		{"warn", (*logger.Logger).Warn, logger.LevelWarn},
		{"error", (*logger.Logger).Error, logger.LevelError},
		{"debug", (*logger.Logger).Debug, logger.LevelDebug},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := logger.New(&buf, "")
			tc.log(l, "hello")
			e := decode(t, &buf)
			if e.Level != tc.wantLvl {
				t.Errorf("level = %q, want %q", e.Level, tc.wantLvl)
			}
			if e.Message != "hello" {
				t.Errorf("msg = %q, want %q", e.Message, "hello")
			}
		})
	}
}

func TestLogger_ProcessField(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, "myapp")
	l.Info("started")
	e := decode(t, &buf)
	if e.Process != "myapp" {
		t.Errorf("process = %q, want %q", e.Process, "myapp")
	}
}

func TestLogger_TimestampPresent(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, "")
	l.Info("ts check")
	e := decode(t, &buf)
	if e.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestLogger_NewlineTerminated(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, "")
	l.Info("nl")
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Error("log entry should end with newline")
	}
}

func TestLogger_NilWriterUsesStdout(t *testing.T) {
	// Just ensure New does not panic with nil writer.
	l := logger.New(nil, "")
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}
