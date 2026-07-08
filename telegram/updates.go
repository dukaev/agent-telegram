// Package telegram provides Telegram client wrapper and update handling.
package telegram

import (
	"fmt"
	"sync"
	"time"

	"agent-telegram/telegram/helpers"
	"agent-telegram/telegram/types"
	"github.com/gotd/td/tg"
)

// UpdateStore stores Telegram updates in memory.
type UpdateStore struct {
	mu       sync.RWMutex
	updates  []types.StoredUpdate
	nextID   int64
	limit    int
	onUpdate func(types.StoredUpdate)
	wg       sync.WaitGroup // tracks in-flight onUpdate callbacks
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

// SetOnUpdate sets a callback that is called after each update is stored.
// The callback is called outside the lock to avoid deadlocks.
func (s *UpdateStore) SetOnUpdate(fn func(types.StoredUpdate)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onUpdate = fn
}

// Add adds a new update to the store.
func (s *UpdateStore) Add(update types.StoredUpdate) {
	s.mu.Lock()

	update.ID = s.nextID
	update.Timestamp = time.Now()
	s.nextID++

	s.updates = append(s.updates, update)

	// Trim old updates if we exceed the limit
	if len(s.updates) > s.limit {
		over := len(s.updates) - s.limit
		s.updates = s.updates[over:]
	}

	onUpdate := s.onUpdate
	s.mu.Unlock()

	if onUpdate != nil {
		s.wg.Go(func() {
			onUpdate(update)
		})
	}
}

// Wait blocks until all in-flight onUpdate callbacks complete.
// Call this during shutdown to prevent goroutine leaks.
func (s *UpdateStore) Wait() {
	s.wg.Wait()
}

// Get returns updates (newest first) without removing them from the store.
// Use offset to get updates after a specific update ID (for polling).
// When offset > 0, only updates with ID > offset are returned.
func (s *UpdateStore) Get(limit int, offset ...int64) []types.StoredUpdate {
	s.mu.RLock()
	defer s.mu.RUnlock()

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

	// Determine starting index based on offset
	startIdx := 0
	if len(offset) > 0 && offset[0] > 0 {
		for i, u := range s.updates {
			if u.ID > offset[0] {
				startIdx = i
				break
			}
			// If we've gone through all updates and none are newer, return empty
			if i == total-1 {
				return []types.StoredUpdate{}
			}
		}
	}

	// Get available updates from startIdx
	available := s.updates[startIdx:]
	count := len(available)
	if count > limit {
		// Take the newest ones
		available = available[count-limit:]
		count = limit
	}

	// Copy and reverse to have newest first
	result := make([]types.StoredUpdate, count)
	copy(result, available)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

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
//
//nolint:nestif // Function requires extracting multiple nested message fields
func MessageData(msg tg.MessageClass, entities tg.Entities) map[string]interface{} {
	data := map[string]interface{}{}

	if m, ok := msg.(*tg.Message); ok {
		data["id"] = m.ID
		data["text"] = m.Message
		data["date"] = m.Date
		data["out"] = m.Out

		// Simplify from_id to string format
		if m.FromID != nil {
			data["from"] = helpers.FormatPeer(m.FromID, helpers.PeerFormatTyped)
			if name := getSenderName(entities, m.FromID); name != "" {
				data["from_name"] = name
			}
		}

		// Simplify peer_id to string format
		if m.PeerID != nil {
			data["peer"] = helpers.FormatPeer(m.PeerID, helpers.PeerFormatTyped)
		}

		// Add media data if present
		if m.Media != nil {
			data["media"] = convertMediaForUpdate(m.Media)
		}

		// Add inline buttons if present
		if m.ReplyMarkup != nil {
			data["buttons"] = extractButtonsData(m.ReplyMarkup)
		}

		// Channel post extra fields
		if views, ok := m.GetViews(); ok {
			data["views"] = views
		}
		if fwds, ok := m.GetForwards(); ok {
			data["forwards"] = fwds
		}
		if gid, ok := m.GetGroupedID(); ok {
			data["grouped_id"] = gid
		}
		if author, ok := m.GetPostAuthor(); ok && author != "" {
			data["post_author"] = author
		}
		if m.Post {
			data["post"] = true
		}
		if fwdFrom, ok := m.GetFwdFrom(); ok {
			if fwdFrom.FromID != nil {
				data["fwd_from"] = helpers.FormatPeer(fwdFrom.FromID, helpers.PeerFormatTyped)
			} else if fromName, nameOk := fwdFrom.GetFromName(); nameOk && fromName != "" {
				data["fwd_from"] = fromName
			}
		}
	}
	return data
}

// convertMediaForUpdate converts message media to a map for updates.
func convertMediaForUpdate(media tg.MessageMediaClass) map[string]any {
	result := make(map[string]any)
	switch m := media.(type) {
	case *tg.MessageMediaPhoto:
		result["type"] = "photo"
		if m.Photo != nil {
			if photo, ok := m.Photo.(*tg.Photo); ok {
				result["photo_id"] = photo.ID
			}
		}
	case *tg.MessageMediaDocument:
		result["type"] = "document"
		if m.Document != nil {
			if doc, ok := m.Document.(*tg.Document); ok {
				result["document_id"] = doc.ID
			}
		}
	case *tg.MessageMediaWebPage:
		result["type"] = "webpage"
		if m.Webpage != nil {
			if wp, ok := m.Webpage.(*tg.WebPage); ok {
				result["url"] = wp.URL
				result["display_url"] = wp.DisplayURL
			}
		}
	case *tg.MessageMediaGeo:
		result["type"] = "geo"
		if gp, ok := m.Geo.(*tg.GeoPoint); ok {
			result["lat"] = gp.Lat
			result["long"] = gp.Long
		}
	case *tg.MessageMediaContact:
		result["type"] = "contact"
		result["phone"] = m.PhoneNumber
		result["first_name"] = m.FirstName
		result["last_name"] = m.LastName
	case *tg.MessageMediaPoll:
		result["type"] = "poll"
	case *tg.MessageMediaDice:
		result["type"] = "dice"
		result["value"] = m.Value
		result["emoticon"] = m.Emoticon
	default:
		result["type"] = "unknown"
	}
	return result
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
//
//nolint:nestif // Function requires multiple nested checks to extract user name
func getSenderName(entities tg.Entities, fromID tg.PeerClass) string {
	if p, ok := fromID.(*tg.PeerUser); ok {
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
