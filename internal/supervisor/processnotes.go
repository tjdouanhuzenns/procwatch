package supervisor

import (
	"fmt"
	"sync"
	"time"
)

// ProcessNote is a timestamped free-text annotation attached to a process.
type ProcessNote struct {
	Process   string    `json:"process"`
	Note      string    `json:"note"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

// ProcessNotes stores ordered notes per process, with a configurable cap.
type ProcessNotes struct {
	mu      sync.RWMutex
	notes   map[string][]ProcessNote
	maxPer  int
}

// NewProcessNotes creates a ProcessNotes store. maxPerProcess limits how many
// notes are retained per process; oldest are evicted when the cap is reached.
func NewProcessNotes(maxPerProcess int) *ProcessNotes {
	if maxPerProcess <= 0 {
		maxPerProcess = 50
	}
	return &ProcessNotes{
		notes:  make(map[string][]ProcessNote),
		maxPer: maxPerProcess,
	}
}

// Add appends a note for the given process.
func (pn *ProcessNotes) Add(process, note, author string) error {
	if process == "" {
		return fmt.Errorf("process name must not be empty")
	}
	if note == "" {
		return fmt.Errorf("note must not be empty")
	}
	entry := ProcessNote{
		Process:   process,
		Note:      note,
		Author:    author,
		CreatedAt: time.Now().UTC(),
	}
	pn.mu.Lock()
	defer pn.mu.Unlock()
	pn.notes[process] = append(pn.notes[process], entry)
	if len(pn.notes[process]) > pn.maxPer {
		pn.notes[process] = pn.notes[process][len(pn.notes[process])-pn.maxPer:]
	}
	return nil
}

// ForProcess returns a copy of all notes for the given process.
func (pn *ProcessNotes) ForProcess(process string) []ProcessNote {
	pn.mu.RLock()
	defer pn.mu.RUnlock()
	src := pn.notes[process]
	out := make([]ProcessNote, len(src))
	copy(out, src)
	return out
}

// All returns every note across all processes.
func (pn *ProcessNotes) All() []ProcessNote {
	pn.mu.RLock()
	defer pn.mu.RUnlock()
	var out []ProcessNote
	for _, notes := range pn.notes {
		out = append(out, notes...)
	}
	return out
}

// Clear removes all notes for the given process.
func (pn *ProcessNotes) Clear(process string) {
	pn.mu.Lock()
	defer pn.mu.Unlock()
	delete(pn.notes, process)
}
