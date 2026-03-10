// Package webhook provides HTTP webhook delivery for Telegram updates.
package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"agent-telegram/telegram/types"
)

const (
	defaultBufferSize  = 256
	defaultRetries     = 3
	defaultRetryDelay  = 2 * time.Second
	defaultHTTPTimeout = 10 * time.Second
)

// Filter determines which updates to forward to the webhook.
type Filter struct {
	// Peer filters by peer string (e.g. "@channel", "channel:123").
	// Empty means no filter.
	Peer string
	// Types filters by update type. Empty means all types.
	Types []string
}

// Sender sends Telegram updates to an HTTP endpoint.
type Sender struct {
	url        string
	httpClient *http.Client
	ch         chan types.StoredUpdate
	retries    int
	retryDelay time.Duration
	filter     Filter
}

// Option configures a Sender.
type Option func(*Sender)

// WithRetries sets the number of delivery retries on failure.
func WithRetries(n int) Option {
	return func(s *Sender) { s.retries = n }
}

// WithRetryDelay sets the base delay between retries (exponential backoff).
func WithRetryDelay(d time.Duration) Option {
	return func(s *Sender) { s.retryDelay = d }
}

// WithFilter sets the update filter.
func WithFilter(f Filter) Option {
	return func(s *Sender) { s.filter = f }
}

// New creates a new Sender that will POST updates to the given URL.
func New(url string, opts ...Option) *Sender {
	s := &Sender{
		url:        url,
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
		ch:         make(chan types.StoredUpdate, defaultBufferSize),
		retries:    defaultRetries,
		retryDelay: defaultRetryDelay,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Send enqueues an update for delivery. Non-blocking: drops the update if the
// internal buffer is full (to avoid blocking the Telegram dispatcher goroutine).
func (s *Sender) Send(update types.StoredUpdate) {
	select {
	case s.ch <- update:
	default:
		slog.Warn("webhook: buffer full, dropping update", "update_id", update.ID, "type", update.Type)
	}
}

// Run reads updates from the internal channel and delivers them to the webhook
// URL. It blocks until ctx is cancelled.
func (s *Sender) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-s.ch:
			if !s.matches(update) {
				continue
			}
			s.deliver(ctx, update)
		}
	}
}

// matches returns true if the update passes the filter.
func (s *Sender) matches(update types.StoredUpdate) bool {
	if len(s.filter.Types) > 0 {
		found := false
		for _, t := range s.filter.Types {
			if string(update.Type) == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if s.filter.Peer == "" {
		return true
	}

	return peerMatchesUpdate(update, s.filter.Peer)
}

// deliver POSTs the update to the webhook URL with retry logic.
func (s *Sender) deliver(ctx context.Context, update types.StoredUpdate) {
	body, err := json.Marshal(update)
	if err != nil {
		slog.Error("webhook: failed to marshal update", "error", err)
		return
	}

	for attempt := 0; attempt <= s.retries; attempt++ {
		if attempt > 0 {
			delay := s.retryDelay * time.Duration(1<<(attempt-1)) // exponential backoff
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
		}

		if err := s.post(ctx, body); err != nil {
			slog.Warn("webhook: delivery failed",
				"attempt", attempt+1,
				"max_attempts", s.retries+1,
				"update_id", update.ID,
				"error", err,
			)
			continue
		}

		slog.Debug("webhook: delivered update", "update_id", update.ID, "type", update.Type)
		return
	}

	slog.Error("webhook: giving up after retries", "update_id", update.ID, "retries", s.retries)
}

// post performs a single HTTP POST with the given body.
func (s *Sender) post(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}

// peerMatchesUpdate checks if the update's peer field matches the filter value.
func peerMatchesUpdate(update types.StoredUpdate, filterPeer string) bool {
	// Normalise: strip leading @
	filter := strings.TrimPrefix(filterPeer, "@")

	msg, ok := update.Data["message"].(map[string]any)
	if !ok {
		return false
	}

	peer, _ := msg["peer"].(string) // e.g. "channel:123"
	if peer == "" {
		return false
	}

	// Exact match or partial case-insensitive match
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
