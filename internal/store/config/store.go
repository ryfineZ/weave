// Package config provides atomic TOML file persistence for Weave's
// user-visible configuration (nodes, chains, rules, settings).
//
// Each sub-domain has its own file. Writes are atomic: data is written to a
// temp file in the same directory, then renamed over the target path so readers
// never see a partial write.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Store manages TOML config files under a single directory.
type Store struct {
	dir string
}

// New creates a Store rooted at dir, creating the directory if needed.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("config store: create dir %q: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

// Load reads and unmarshals the TOML file name (e.g. "nodes.toml") into dst.
// If the file does not exist, dst is left unchanged and no error is returned.
func (s *Store) Load(name string, dst any) error {
	path := filepath.Join(s.dir, name)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("config store: read %q: %w", name, err)
	}
	if err := toml.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("config store: parse %q: %w", name, err)
	}
	return nil
}

// Save marshals src as TOML and atomically writes it to name.
func (s *Store) Save(name string, src any) error {
	data, err := toml.Marshal(src)
	if err != nil {
		return fmt.Errorf("config store: marshal %q: %w", name, err)
	}

	target := filepath.Join(s.dir, name)

	// Write to a temp file in the same directory so the rename is atomic.
	tmp, err := os.CreateTemp(s.dir, ".tmp-"+name+"-*")
	if err != nil {
		return fmt.Errorf("config store: create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op if rename succeeded

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("config store: write temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("config store: sync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("config store: close temp: %w", err)
	}
	if err := os.Rename(tmpName, target); err != nil {
		return fmt.Errorf("config store: rename to %q: %w", target, err)
	}
	return nil
}

// Dir returns the root directory of the store.
func (s *Store) Dir() string { return s.dir }
