// Package telegram provides Telegram client wrapper and update handling.
package telegram

import (
	"fmt"
	"sync"
	"time"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// UpdateStore stores Telegram updates in memory.
type UpdateStore struct {
	mu      sync.RWMutex
	updates []types.StoredUpdate
	nextID  int64
	limit   int
}

// NewUpdateStore creates a new UpdateStore with the given limit.
func NewUpdateStore(limit int) *UpdateStore {
	if limit <= 0 {
		limit = 1000
	}
	return &UpdateStore{
		updates: make([]types.StoredUpdate, 0, limit),
		nextID:  1,
		limit:   limit,
	}
}

// Add adds a new update to the store.
func (s *UpdateStore) Add(update types.StoredUpdate) {
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
func (s *UpdateStore) Get(limit int) []types.StoredUpdate {
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
		return []types.StoredUpdate{}
	}

	// Determine how many to return
	count := limit
	if count > total {
		count = total
	}

	// Take from the end (newest)
	start := total - count
	result := make([]types.StoredUpdate, count)
	copy(result, s.updates[start:])

	// Reverse to have newest first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	// Remove returned updates from store
	s.updates = s.updates[:start]

	return result
}

// NewStoredUpdate creates a new types.StoredUpdate from raw data.
func NewStoredUpdate(updateType types.UpdateType, data map[string]interface{}) types.StoredUpdate {
	return types.StoredUpdate{
		Type: updateType,
		Data: data,
	}
}

// MessageData extracts message data from tg.MessageClass.
func MessageData(msg tg.MessageClass, entities tg.Entities) map[string]interface{} {
	data := map[string]interface{}{}

	if m, ok := msg.(*tg.Message); ok {
		data["id"] = m.ID
		data["text"] = m.Message
		data["date"] = m.Date
		data["out"] = m.Out
		if m.FromID != nil {
			data["from_id"] = m.FromID
			// Add sender name
			data["from_name"] = getSenderName(entities, m.FromID)
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
		// Add inline buttons if present
		if m.ReplyMarkup != nil {
			data["buttons"] = extractButtonsData(m.ReplyMarkup)
		}
	}
	return data
}

// extractButtonsData extracts button data from ReplyMarkup.
func extractButtonsData(markup tg.ReplyMarkupClass) []map[string]interface{} {
	rm, ok := markup.(*tg.ReplyInlineMarkup)
	if !ok {
		return nil
	}

	var result []map[string]interface{}
	for _, row := range rm.Rows {
		for _, button := range row.Buttons {
			btnData := map[string]interface{}{"index": len(result)}
			switch b := button.(type) {
			case *tg.KeyboardButtonURL:
				btnData["text"] = b.Text
				btnData["type"] = "url"
				btnData["data"] = b.URL
			case *tg.KeyboardButtonCallback:
				btnData["text"] = b.Text
				btnData["type"] = "callback"
				btnData["data"] = string(b.Data)
			case *tg.KeyboardButtonSwitchInline:
				btnData["text"] = b.Text
				btnData["type"] = "switch_inline"
				btnData["data"] = b.Query
			case *tg.KeyboardButtonGame:
				btnData["text"] = b.Text
				btnData["type"] = "game"
			case *tg.KeyboardButtonBuy:
				btnData["text"] = b.Text
				btnData["type"] = "buy"
			case *tg.KeyboardButtonURLAuth:
				btnData["text"] = b.Text
				btnData["type"] = "url_auth"
				btnData["data"] = b.URL
			}
			result = append(result, btnData)
		}
	}
	return result
}

// getSenderName gets the name of a sender from their peer ID.
func getSenderName(entities tg.Entities, fromID tg.PeerClass) string {
	switch p := fromID.(type) {
	case *tg.PeerUser:
		if user, ok := entities.Users[p.UserID]; ok {
			name := user.FirstName
			if user.LastName != "" {
				name += " " + user.LastName
			}
			if name == "" && user.Username != "" {
				name = "@" + user.Username
			}
			if name == "" {
				name = fmt.Sprintf("user:%d", user.ID)
			}
			if user.Bot {
				name += " (bot)"
			}
			return name
		}
	}
	return ""
}
