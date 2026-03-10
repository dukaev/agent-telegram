package callback

import (
	"context"
	"log/slog"
	"strings"
	"sync/atomic"

	"agent-telegram/internal/webhook"
	"agent-telegram/telegram/types"
)

// eventTypeToUpdateType maps API event types to internal update types.
var eventTypeToUpdateType = map[string]string{
	"new_post":  string(types.UpdateTypeNewMessage),
	"edit_post": string(types.UpdateTypeEditMessage),
}

// Manager routes Telegram updates to matching subscriptions and delivers them
// to the configured callback URL.
type Manager struct {
	store   *Store
	sender  *webhook.Sender
	pending atomic.Int64
}

// NewManager creates a Manager using the given store.
// The sender is created lazily when a callback URL is set.
func NewManager(store *Store) *Manager {
	m := &Manager{store: store}
	m.rebuildSender()
	return m
}

// rebuildSender recreates the sender from the current store state.
// Call after changing the callback URL.
func (m *Manager) rebuildSender() {
	state := m.store.Get()
	if state.CallbackURL == "" || !state.Verified {
		m.sender = nil
		return
	}
	m.sender = webhook.New(
		state.CallbackURL,
		webhook.WithRetries(3),
	)
}

// Run starts the sender goroutine (if a verified callback URL exists).
// It blocks until ctx is cancelled.
func (m *Manager) Run(ctx context.Context) {
	if m.sender != nil {
		m.sender.Run(ctx)
	} else {
		<-ctx.Done()
	}
}

// HandleUpdate is the callback for UpdateStore.SetOnUpdate.
// It checks all active subscriptions and dispatches matching updates.
func (m *Manager) HandleUpdate(update types.StoredUpdate) {
	if m.sender == nil {
		return
	}

	state := m.store.Get()
	for _, sub := range state.Subscriptions {
		if m.matchesSubscription(update, sub) {
			m.pending.Add(1)
			m.sender.Send(update)
			m.pending.Add(-1)
			return // send once even if multiple subs match
		}
	}
}

// PendingCount returns the number of updates currently queued for delivery.
func (m *Manager) PendingCount() int64 {
	return m.pending.Load()
}

// SetCallbackURL sets a new callback URL (triggers verification flow).
// Returns the verify code to expect from the target URL.
func (m *Manager) SetCallbackURL(url string) (string, error) {
	code, err := m.store.SetCallbackURL(url)
	if err != nil {
		return "", err
	}
	m.sender = nil // invalidate until verified
	return code, nil
}

// ConfirmVerification marks the URL as verified and rebuilds the sender.
func (m *Manager) ConfirmVerification() error {
	if err := m.store.MarkVerified(); err != nil {
		return err
	}
	m.rebuildSender()
	if m.sender != nil {
		slog.Info("callback: URL verified and active", "url", m.store.Get().CallbackURL)
	}
	return nil
}

// matchesSubscription returns true if the update matches the subscription's
// channel and event type filters.
func (m *Manager) matchesSubscription(update types.StoredUpdate, sub Subscription) bool {
	// Check event type
	matched := false
	for _, et := range sub.EventTypes {
		if updateType, ok := eventTypeToUpdateType[et]; ok {
			if string(update.Type) == updateType {
				matched = true
				break
			}
		}
	}
	if !matched {
		return false
	}

	// Check peer
	if sub.ChannelID == "" {
		return true
	}
	return peerMatchesUpdate(update, sub.ChannelID)
}

// peerMatchesUpdate checks if the update's peer matches the subscription channel.
func peerMatchesUpdate(update types.StoredUpdate, filterPeer string) bool {
	filter := strings.TrimPrefix(filterPeer, "@")

	msg, ok := update.Data["message"].(map[string]any)
	if !ok {
		return false
	}

	peer, _ := msg["peer"].(string) // e.g. "channel:123"
	if peer == "" {
		return false
	}

	if strings.EqualFold(peer, filter) {
		return true
	}
	if strings.Contains(strings.ToLower(peer), strings.ToLower(filter)) {
		return true
	}

	// Numeric ID match: "channel:123" vs "123"
	parts := strings.SplitN(peer, ":", 2)
	if len(parts) == 2 && parts[1] == filter {
		return true
	}

	return false
}
