package sessionstore

import (
	"context"
	"errors"
	"sync"

	"github.com/gotd/td/session"
)

// Storage adapts a pluggable Store to gotd's session.Storage interface.
type Storage struct {
	store   Store
	profile string
	mu      sync.Mutex
}

var _ session.Storage = (*Storage)(nil)

func NewStorage(store Store, profile string) *Storage {
	return &Storage{store: store, profile: profile}
}

func (s *Storage) LoadSession(ctx context.Context) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := s.store.Load(ctx, s.profile)
	if errors.Is(err, ErrNotFound) {
		return nil, session.ErrNotFound
	}
	return data, err
}

func (s *Storage) StoreSession(ctx context.Context, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.store.Save(ctx, s.profile, data)
}

// ExportSession reads a copy of the current opaque session bytes.
func (s *Storage) ExportSession() []byte {
	data, err := s.LoadSession(context.Background())
	if err != nil {
		return nil
	}
	return data
}

// Clear removes the selected profile. It is intentionally compatible with the
// existing Telegram client lifecycle interface.
func (s *Storage) Clear() {
	_ = s.Delete(context.Background())
}

// ClearSession removes the selected profile and reports provider failures.
func (s *Storage) ClearSession(ctx context.Context) error {
	return s.Delete(ctx)
}

// Delete removes the selected profile and returns any provider error.
func (s *Storage) Delete(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.store.Delete(ctx, s.profile)
}

func (s *Storage) Provider() string { return s.store.Provider() }
func (s *Storage) Profile() string  { return s.profile }
func (s *Storage) Persistent() bool { return s.store.Persistent() }

func (s *Storage) Selection() Selection {
	return Selection{Provider: s.Provider(), Profile: s.Profile(), Persistent: s.Persistent()}
}
