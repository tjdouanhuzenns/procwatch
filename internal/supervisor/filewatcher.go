package supervisor

import (
	"context"
	"os"
	"sync"
	"time"
)

// FileWatcherConfig holds configuration for the file watcher.
type FileWatcherConfig struct {
	PollInterval time.Duration
}

// DefaultFileWatcherConfig returns sensible defaults.
func DefaultFileWatcherConfig() FileWatcherConfig {
	return FileWatcherConfig{
		PollInterval: 2 * time.Second,
	}
}

// FileEvent describes a change detected on a watched file.
type FileEvent struct {
	Path    string
	Process string
	ModTime time.Time
}

// fileEntry tracks a single watched file.
type fileEntry struct {
	process string
	modTime time.Time
}

// FileWatcher polls registered files for modification and emits events.
type FileWatcher struct {
	mu      sync.Mutex
	cfg     FileWatcherConfig
	files   map[string]*fileEntry // path -> entry
	events  chan FileEvent
}

// NewFileWatcher creates a FileWatcher with the given config.
func NewFileWatcher(cfg FileWatcherConfig) *FileWatcher {
	return &FileWatcher{
		cfg:   cfg,
		files: make(map[string]*fileEntry),
		events: make(chan FileEvent, 64),
	}
}

// Watch registers a file path associated with a process name.
func (fw *FileWatcher) Watch(process, path string) error {
	if process == "" {
		return errEmptyName
	}
	if path == "" {
		return errEmptyCommand
	}
	info, err := os.Stat(path)
	var modTime time.Time
	if err == nil {
		modTime = info.ModTime()
	}
	fw.mu.Lock()
	fw.files[path] = &fileEntry{process: process, modTime: modTime}
	fw.mu.Unlock()
	return nil
}

// Unwatch removes a file from the watch list.
func (fw *FileWatcher) Unwatch(path string) {
	fw.mu.Lock()
	delete(fw.files, path)
	fw.mu.Unlock()
}

// Events returns the channel on which file change events are delivered.
func (fw *FileWatcher) Events() <-chan FileEvent {
	return fw.events
}

// Run starts polling until ctx is cancelled.
func (fw *FileWatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(fw.cfg.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			close(fw.events)
			return
		case <-ticker.C:
			fw.poll()
		}
	}
}

func (fw *FileWatcher) poll() {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	for path, entry := range fw.files {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.ModTime().After(entry.modTime) {
			entry.modTime = info.ModTime()
			select {
			case fw.events <- FileEvent{Path: path, Process: entry.process, ModTime: info.ModTime()}:
			default:
			}
		}
	}
}
