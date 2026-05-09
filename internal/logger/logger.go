package logger

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

// Level represents a log severity level.
type Level string

const (
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
	LevelDebug Level = "debug"
)

// Entry is a structured log record.
type Entry struct {
	Timestamp string `json:"ts"`
	Level     Level  `json:"level"`
	Process   string `json:"process,omitempty"`
	Message   string `json:"msg"`
}

// Logger writes structured JSON log entries to an io.Writer.
type Logger struct {
	writer  io.Writer
	process string
}

// New creates a Logger that writes to w.
// If w is nil, os.Stdout is used.
func New(w io.Writer, process string) *Logger {
	if w == nil {
		w = os.Stdout
	}
	return &Logger{writer: w, process: process}
}

func (l *Logger) write(level Level, msg string) {
	e := Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Process:   l.process,
		Message:   msg,
	}
	data, err := json.Marshal(e)
	if err != nil {
		return
	}
	_, _ = l.writer.Write(append(data, '\n'))
}

// Info logs an informational message.
func (l *Logger) Info(msg string) { l.write(LevelInfo, msg) }

// Warn logs a warning message.
func (l *Logger) Warn(msg string) { l.write(LevelWarn, msg) }

// Error logs an error message.
func (l *Logger) Error(msg string) { l.write(LevelError, msg) }

// Debug logs a debug message.
func (l *Logger) Debug(msg string) { l.write(LevelDebug, msg) }
