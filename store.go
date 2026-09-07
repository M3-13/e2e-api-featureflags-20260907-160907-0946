package main

import (
	"errors"
	"sync"
)

// ErrExists is returned by Store.Create when a feature with the same key
// already exists.
var ErrExists = errors.New("feature already exists")

// Feature is a single feature flag.
type Feature struct {
	Key            string `json:"key"`
	Enabled        bool   `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent int    `json:"rollout_percent"`
}

// Store is a thread-safe in-memory feature flag store.
type Store struct {
	mu    sync.RWMutex
	flags map[string]Feature
}

// NewStore returns an empty, ready-to-use Store.
func NewStore() *Store {
	return &Store{flags: make(map[string]Feature)}
}

// Create adds a new feature flag. It returns ErrExists if a flag with the
// same key already exists.
func (s *Store) Create(f Feature) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[f.Key]; ok {
		return ErrExists
	}
	s.flags[f.Key] = f
	return nil
}

// List returns all feature flags in no particular order.
func (s *Store) List() []Feature {
	s.mu.RLock()
	defer s.mu.RUnlock()
	flags := make([]Feature, 0, len(s.flags))
	for _, f := range s.flags {
		flags = append(flags, f)
	}
	return flags
}

// Get returns the feature flag for key and whether it was found.
func (s *Store) Get(key string) (Feature, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.flags[key]
	return f, ok
}

// Update replaces the feature flag identified by key. The key is immutable:
// it is taken from the key argument, not from f. It returns the stored flag
// and whether the key existed.
func (s *Store) Update(key string, f Feature) (Feature, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[key]; !ok {
		return Feature{}, false
	}
	f.Key = key
	s.flags[key] = f
	return f, true
}

// Delete removes the feature flag identified by key and reports whether it
// existed.
func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[key]; !ok {
		return false
	}
	delete(s.flags, key)
	return true
}
