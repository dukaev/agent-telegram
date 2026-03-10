// Package webhook provides HTTP webhook delivery for Telegram updates.
package webhook

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	defaultBufferSize  = 256
	defaultRetries     = 3
	defaultRetryDelay  = 2 * time.Second
	defaultHTTPTimeout = 10 * time.Second
	drainTimeout       = 5 * time.Second
)

// Sender delivers pre-marshaled JSON payloads to an HTTP endpoint with retry.
type Sender struct {
	url        string
	httpClient *http.Client
	ch         chan []byte
	retries    int
	retryDelay time.Duration
	onError    func(string)
	dropped    atomic.Int64
	done       chan struct{} // closed when Run() exits
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

// WithOnError sets a callback invoked after all retries are exhausted.
func WithOnError(fn func(string)) Option {
	return func(s *Sender) { s.onError = fn }
}

// New creates a new Sender that will POST payloads to the given URL.
func New(url string, opts ...Option) *Sender {
	s := &Sender{
		url:        url,
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
		ch:         make(chan []byte, defaultBufferSize),
		retries:    defaultRetries,
		retryDelay: defaultRetryDelay,
		done:       make(chan struct{}),
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Send enqueues a pre-marshaled JSON payload for delivery. Non-blocking: drops
// the payload if the internal buffer is full.
func (s *Sender) Send(payload []byte) {
	select {
	case s.ch <- payload:
	default:
		n := s.dropped.Add(1)
		slog.Warn("webhook: buffer full, dropping payload", "total_dropped", n)
	}
}

// QueueLen returns the number of payloads currently waiting in the buffer.
func (s *Sender) QueueLen() int {
	return len(s.ch)
}

// DroppedCount returns the total number of payloads dropped due to full buffer.
func (s *Sender) DroppedCount() int64 {
	return s.dropped.Load()
}

// Done returns a channel that is closed when Run() has fully exited
// (including draining the queue).
func (s *Sender) Done() <-chan struct{} {
	return s.done
}

// Run reads payloads from the internal channel and delivers them to the webhook
// URL. When ctx is cancelled, it drains remaining queued payloads before exiting.
func (s *Sender) Run(ctx context.Context) {
	defer close(s.done)

	for {
		select {
		case <-ctx.Done():
			s.drain() //nolint:contextcheck // drain uses fresh background context since parent is cancelled
			return
		case payload := <-s.ch:
			s.deliver(ctx, payload)
		}
	}
}

// drain delivers all remaining payloads in the buffer with a timeout.
func (s *Sender) drain() {
	remaining := len(s.ch)
	if remaining == 0 {
		return
	}

	slog.Info("webhook: draining queue", "remaining", remaining)

	drainCtx, cancel := context.WithTimeout(context.Background(), drainTimeout)
	defer cancel()

	for {
		select {
		case payload := <-s.ch:
			s.deliver(drainCtx, payload)
		case <-drainCtx.Done():
			lost := len(s.ch)
			if lost > 0 {
				slog.Warn("webhook: drain timeout, lost payloads", "count", lost)
			}
			return
		default:
			slog.Info("webhook: queue drained")
			return
		}
	}
}

// deliver POSTs the payload to the webhook URL with retry logic.
func (s *Sender) deliver(ctx context.Context, payload []byte) {
	var lastErr error
	for attempt := 0; attempt <= s.retries; attempt++ {
		if attempt > 0 {
			delay := s.retryDelay * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
		}

		if err := s.post(ctx, payload); err != nil {
			lastErr = err
			slog.Warn("webhook: delivery failed",
				"attempt", attempt+1,
				"max_attempts", s.retries+1,
				"error", err,
			)
			continue
		}

		slog.Debug("webhook: delivered payload")
		return
	}

	msg := fmt.Sprintf("giving up after %d retries: %v", s.retries, lastErr)
	slog.Error("webhook: " + msg)
	if s.onError != nil {
		s.onError(msg)
	}
}

// post performs a single HTTP POST with the given body.
func (s *Sender) post(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, s.url, bytes.NewReader(body),
	)
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
