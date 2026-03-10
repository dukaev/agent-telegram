package callback

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"

	"agent-telegram/internal/webhook"
	"agent-telegram/telegram/types"
)

// PeerResolver resolves and normalizes a peer string (e.g., "@username" or "-100123")
// to a typed ID string (e.g., "channel:123"). Returns an error if the peer doesn't exist.
type PeerResolver func(ctx context.Context, peer string) (string, error)

// eventTypeToUpdateType maps API event types to internal update types.
var eventTypeToUpdateType = map[string]string{
	"new_post":    string(types.UpdateTypeNewMessage),
	"edit_post":   string(types.UpdateTypeEditMessage),
	"delete_post": string(types.UpdateTypeDelete),
}

// Manager routes Telegram updates to matching subscriptions and delivers them
// to the configured callback URL.
type Manager struct {
	store        *Store
	mu           sync.Mutex
	sender       *webhook.Sender
	senderCancel context.CancelFunc
	parentCtx    context.Context  //nolint:containedctx // intentionally stored for goroutine lifecycle
	peerResolver PeerResolver
}

// NewManager creates a Manager using the given store.
func NewManager(store *Store) *Manager {
	return &Manager{store: store}
}

// WithPeerResolver sets the function used to validate and normalize channel IDs.
func (m *Manager) WithPeerResolver(fn PeerResolver) {
	m.mu.Lock()
	m.peerResolver = fn
	m.mu.Unlock()
}

// ResolveChannelID validates and normalizes a channel ID via Telegram API.
// Returns the normalized ID (e.g., "channel:123") or the original if no resolver is set.
func (m *Manager) ResolveChannelID(ctx context.Context, channelID string) (string, error) {
	m.mu.Lock()
	resolver := m.peerResolver
	m.mu.Unlock()
	if resolver == nil {
		return channelID, nil
	}
	return resolver(ctx, channelID)
}

// Run stores the parent context and starts the sender goroutine if a verified
// URL already exists. It blocks until ctx is cancelled.
func (m *Manager) Run(ctx context.Context) {
	m.mu.Lock()
	m.parentCtx = ctx
	m.mu.Unlock()

	m.rebuildSender() //nolint:contextcheck // parentCtx is stored explicitly for sender goroutine lifecycle

	<-ctx.Done()
}

// rebuildSender stops any existing sender goroutine, creates a new Sender
// for the current verified URL, and starts its goroutine if Run() is active.
func (m *Manager) rebuildSender() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.senderCancel != nil {
		m.senderCancel()
		m.senderCancel = nil
	}
	m.sender = nil

	state := m.store.Get()
	if state.CallbackURL == "" || !state.Verified {
		return
	}

	m.sender = webhook.New(
		state.CallbackURL,
		webhook.WithRetries(3),
		webhook.WithOnError(func(msg string) { m.store.RecordError(msg) }),
	)

	if m.parentCtx != nil {
		sCtx, cancel := context.WithCancel(m.parentCtx)
		m.senderCancel = cancel
		go m.sender.Run(sCtx)
		slog.Info("callback: sender started", "url", state.CallbackURL)
	}
}

// HandleUpdate is the callback for UpdateStore.SetOnUpdate.
// It checks all active subscriptions and dispatches matching updates.
func (m *Manager) HandleUpdate(update types.StoredUpdate) {
	m.mu.Lock()
	sender := m.sender
	m.mu.Unlock()

	if sender == nil {
		return
	}

	state := m.store.Get()
	for _, sub := range state.Subscriptions {
		eventType, ok := m.matchesSubscription(update, sub)
		if !ok {
			continue
		}
		event := toCallbackEvent(update, eventType)
		payload, err := json.Marshal(event)
		if err != nil {
			slog.Error("callback: marshal event failed", "error", err)
			return
		}
		sender.Send(payload)
		return // send once even if multiple subs match
	}
}

// PendingCount returns the number of updates currently queued for delivery.
func (m *Manager) PendingCount() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sender == nil {
		return 0
	}
	return int64(m.sender.QueueLen())
}

// SetCallbackURL sets a new callback URL (triggers verification flow).
// Returns the verify code to expect from the target URL.
func (m *Manager) SetCallbackURL(url string) (string, error) {
	code, err := m.store.SetCallbackURL(url)
	if err != nil {
		return "", err
	}
	// Invalidate sender until verification completes.
	m.mu.Lock()
	if m.senderCancel != nil {
		m.senderCancel()
		m.senderCancel = nil
	}
	m.sender = nil
	m.mu.Unlock()
	return code, nil
}

// ConfirmVerification marks the URL as verified and starts the new sender.
func (m *Manager) ConfirmVerification() error {
	if err := m.store.MarkVerified(); err != nil {
		return err
	}
	m.rebuildSender()
	return nil
}

// matchesSubscription returns the matched event type and true if the update
// matches the subscription's channel and event type filters.
func (m *Manager) matchesSubscription(
	update types.StoredUpdate, sub Subscription,
) (string, bool) {
	for _, et := range sub.EventTypes {
		updateType, ok := eventTypeToUpdateType[et]
		if !ok {
			continue
		}
		if string(update.Type) != updateType {
			continue
		}
		// Event type matched — check peer
		if sub.ChannelID == "" || peerMatchesUpdate(update, sub.ChannelID) {
			return et, true
		}
	}
	return "", false
}

// toCallbackEvent transforms a StoredUpdate into an Event.
//nolint:nestif // Extracting multiple optional fields from a map requires nested type assertions
func toCallbackEvent(update types.StoredUpdate, eventType string) Event {
	msg, _ := update.Data["message"].(map[string]any)
	post := Post{IsDeleted: eventType == "delete_post"}
	if msg != nil {
		if id, ok := msg["id"].(int); ok {
			post.ID = id
		}
		if date, ok := msg["date"].(int); ok {
			post.Date = date
		}
		if text, ok := msg["text"].(string); ok {
			post.Text = text
		}
		if views, ok := msg["views"].(int); ok {
			post.Views = views
		}
		if peer, ok := msg["peer"].(string); ok {
			post.ChannelID = peer
		}
		if fwdFrom, ok := msg["fwd_from"].(string); ok {
			post.ForwardedFrom = fwdFrom
		}
		if gid, ok := msg["grouped_id"].(int64); ok {
			post.GroupID = &gid
		}
		if author, ok := msg["post_author"].(string); ok {
			post.PostAuthor = author
		}
		if media, ok := msg["media"].(map[string]any); ok {
			post.Media = media
		}
	}
	return Event{EventType: eventType, Post: post}
}

// peerMatchesUpdate checks if the update's peer matches the subscription channel.
// Supports exact match (case-insensitive) and numeric ID match ("channel:123" vs "123").
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

	// Exact match (case-insensitive)
	if strings.EqualFold(peer, filter) {
		return true
	}

	// Numeric ID match: "channel:123" vs "123"
	parts := strings.SplitN(peer, ":", 2)
	if len(parts) == 2 && parts[1] == filter {
		return true
	}

	return false
}
