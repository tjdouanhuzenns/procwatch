package supervisor

import (
	"fmt"
	"sync"
	"time"
)

// AuditEntry records a single command issued to a process.
type AuditEntry struct {
	Process   string    `json:"process"`
	Command   string    `json:"command"`
	IssuedBy  string    `json:"issued_by"`
	Timestamp time.Time `json:"timestamp"`
}

// CommandAudit stores a bounded history of commands issued to processes.
type CommandAudit struct {
	mu      sync.RWMutex
	entries []AuditEntry
	maxSize int
}

// NewCommandAudit creates a CommandAudit with the given maximum history size.
func NewCommandAudit(maxSize int) *CommandAudit {
	if maxSize <= 0 {
		maxSize = 256
	}
	return &CommandAudit{
		entries: make([]AuditEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Record appends a command entry, evicting the oldest when full.
func (a *CommandAudit) Record(process, command, issuedBy string) error {
	if process == "" {
		return fmt.Errorf("commandaudit: process name must not be empty")
	}
	if command == "" {
		return fmt.Errorf("commandaudit: command must not be empty")
	}
	entry := AuditEntry{
		Process:   process,
		Command:   command,
		IssuedBy:  issuedBy,
		Timestamp: time.Now().UTC(),
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.entries) >= a.maxSize {
		a.entries = a.entries[1:]
	}
	a.entries = append(a.entries, entry)
	return nil
}

// All returns a copy of all audit entries in chronological order.
func (a *CommandAudit) All() []AuditEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]AuditEntry, len(a.entries))
	copy(out, a.entries)
	return out
}

// ForProcess returns audit entries for a specific process.
func (a *CommandAudit) ForProcess(process string) []AuditEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	var out []AuditEntry
	for _, e := range a.entries {
		if e.Process == process {
			out = append(out, e)
		}
	}
	return out
}

// Len returns the number of recorded entries.
func (a *CommandAudit) Len() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.entries)
}
