package sessionstore

import (
	"context"
	"sync"
)

const MemoryProvider = "memory"

type memoryStore struct {
	mu   sync.Mutex
	data map[string][]byte
}

var processMemoryStore = &memoryStore{data: make(map[string][]byte)}

func init() {
	RegisterProvider(MemoryProvider, func() (Store, error) {
		return processMemoryStore, nil
	})
}

func (s *memoryStore) Provider() string { return MemoryProvider }
func (s *memoryStore) Persistent() bool { return false }

func (s *memoryStore) Load(_ context.Context, profile string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data := s.data[profile]
	if len(data) == 0 {
		return nil, ErrNotFound
	}
	return append([]byte(nil), data...), nil
}

func (s *memoryStore) Save(_ context.Context, profile string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[profile] = append([]byte(nil), data...)
	return nil
}

func (s *memoryStore) Delete(_ context.Context, profile string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, profile)
	return nil
}
