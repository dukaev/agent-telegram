// Package telegram provides Telegram client wrapper and update handling.
package telegram

import (
	"sync"
	"time"

	"github.com/gotd/td/tg"
)

// UpdateType represents the type of Telegram update.
type UpdateType string

const (
	UpdateTypeNewMessage  UpdateType = "new_message"
	UpdateTypeEditMessage UpdateType = "edit_message"
	UpdateTypeNewChat     UpdateType = "new_chat"
	UpdateTypeDelete      UpdateType = "delete"
	UpdateTypeOther       UpdateType = "other"
)

// StoredUpdate represents a stored Telegram update.
type StoredUpdate struct {
	ID        int64                      `json:"id"`
	Type      UpdateType                 `json:"type"`
	Timestamp time.Time                  `json:"timestamp"`
	Data      map[string]interface{}     `json:"data"`
}

// UpdateStore stores Telegram updates in memory.
type UpdateStore struct {
	mu      sync.RWMutex
	updates []StoredUpdate
	nextID  int64
	limit   int
}

// NewUpdateStore creates a new UpdateStore with the given limit.
func NewUpdateStore(limit int) *UpdateStore {
	if limit <= 0 {
		limit = 1000
	}
	return &UpdateStore{
		updates: make([]StoredUpdate, 0, limit),
		nextID:  1,
		limit:   limit,
	}
}

// Add adds a new update to the store.
func (s *UpdateStore) Add(update StoredUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	update.ID = s.nextID
	update.Timestamp = time.Now()
	s.nextID++

	s.updates = append(s.updates, update)

	// Trim old updates if we exceed the limit
	if len(s.updates) > s.limit {
		// Remove oldest updates (from the beginning)
		over := len(s.updates) - s.limit
		s.updates = s.updates[over:]
	}
}

// Get pops and returns updates (newest first, removes from store).
func (s *UpdateStore) Get(limit int) []StoredUpdate {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set defaults
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	total := len(s.updates)
	if total == 0 {
		return []StoredUpdate{}
	}

	// Determine how many to return
	count := limit
	if count > total {
		count = total
	}

	// Take from the end (newest)
	start := total - count
	result := make([]StoredUpdate, count)
	copy(result, s.updates[start:])

	// Reverse to have newest first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	// Remove returned updates from store
	s.updates = s.updates[:start]

	return result
}

// NewStoredUpdate creates a new StoredUpdate from raw data.
func NewStoredUpdate(updateType UpdateType, data map[string]interface{}) StoredUpdate {
	return StoredUpdate{
		Type: updateType,
		Data: data,
	}
}

// MessageData extracts message data from tg.MessageClass.
func MessageData(msg tg.MessageClass) map[string]interface{} {
	data := map[string]interface{}{}

	switch m := msg.(type) {
	case *tg.Message:
		data["id"] = m.ID
		data["text"] = m.Message
		data["date"] = m.Date
		data["out"] = m.Out
		if m.FromID != nil {
			data["from_id"] = m.FromID
		}
		if m.PeerID != nil {
			data["peer_id"] = m.PeerID
		}
		if m.ReplyTo != nil {
			data["reply_to"] = m.ReplyTo
		}
		if m.Media != nil {
			data["media"] = m.Media
		}
	}
	return data
}
