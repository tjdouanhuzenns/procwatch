package logger

import (
	"io"
	"sync"
)

// multiWriter fans out writes to multiple io.Writers, similar to io.MultiWriter
// but continues on partial errors and is safe for concurrent use.
type multiWriter struct {
	mu      sync.Mutex
	writers []io.Writer
}

// MultiWriter returns an io.Writer that duplicates writes to all provided writers.
func MultiWriter(writers ...io.Writer) io.Writer {
	w := make([]io.Writer, len(writers))
	copy(w, writers)
	return &multiWriter{writers: w}
}

func (m *multiWriter) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, w := range m.writers {
		if _, werr := w.Write(p); werr != nil && err == nil {
			err = werr
		}
	}
	return len(p), err
}

// NewMulti creates a Logger that writes to all provided writers.
func NewMulti(process string, writers ...io.Writer) *Logger {
	return New(MultiWriter(writers...), process)
}
