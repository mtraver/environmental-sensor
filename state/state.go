package state

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	stateDirEnvVar  = "STATE_DIRECTORY"
	defaultStateDir = "/var/lib/iotcorelogger"
)

// Store persists and retrieves opaque state as bytes.
type Store interface {
	// Save persists state, overwriting any previously saved state.
	Save(state []byte) error

	// Load returns previously saved state. It returns nil, nil if nothing has
	// been saved yet.
	Load() ([]byte, error)
}

// FileStore is a Store backed by a file on disk.
type FileStore struct {
	path string
}

// NewFileStore returns a FileStore that persists state to a file named name
// inside $STATE_DIRECTORY. If $STATE_DIRECTORY is unset, it falls back to
// /var/lib/iotcorelogger.
func NewFileStore(name string) (*FileStore, error) {
	dir := os.Getenv(stateDirEnvVar)
	if dir == "" {
		dir = defaultStateDir
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("state: failed to create state directory %q: %w", dir, err)
	}

	return &FileStore{path: filepath.Join(dir, name)}, nil
}

// Save writes state to disk atomically. It writes to a temp file in the same
// directory, then renames it over the destination. This avoids leaving a
// corrupt/truncated file in place if the process is interrupted mid-write.
func (f *FileStore) Save(state []byte) error {
	dir := filepath.Dir(f.path)

	tmp, err := os.CreateTemp(dir, filepath.Base(f.path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("state: failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(state); err != nil {
		tmp.Close()
		return fmt.Errorf("state: failed to write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("state: failed to sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("state: failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, f.path); err != nil {
		return fmt.Errorf("state: failed to rename temp file into place: %w", err)
	}

	return nil
}

// Load reads previously saved state. It returns (nil, nil) if no state has been saved yet.
func (f *FileStore) Load() ([]byte, error) {
	b, err := os.ReadFile(f.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("state: failed to read state file: %w", err)
	}

	return b, nil
}
