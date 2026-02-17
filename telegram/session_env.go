package telegram

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/gotd/td/session"
)

// EnvStorage implements session.Storage, keeping the session in memory.
// Used when session data is loaded from an environment variable.
type EnvStorage struct {
	data []byte
	mu   sync.Mutex
}

// Compile-time check that EnvStorage implements session.Storage.
var _ session.Storage = (*EnvStorage)(nil)

// NewEnvStorage creates a new EnvStorage by decoding a base64-encoded session string.
func NewEnvStorage(base64str string) (*EnvStorage, error) {
	data, err := base64.StdEncoding.DecodeString(base64str)
	if err != nil {
		return nil, fmt.Errorf("failed to decode TELEGRAM_SESSION: %w", err)
	}
	return &EnvStorage{data: data}, nil
}

// LoadSession returns the session data from memory.
func (s *EnvStorage) LoadSession(_ context.Context) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.data) == 0 {
		return nil, session.ErrNotFound
	}
	// Return a copy to prevent mutation
	out := make([]byte, len(s.data))
	copy(out, s.data)
	return out, nil
}

// StoreSession updates the session data in memory.
// Telegram periodically re-saves the session; this keeps it in sync.
func (s *EnvStorage) StoreSession(_ context.Context, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make([]byte, len(data))
	copy(s.data, data)
	return nil
}
