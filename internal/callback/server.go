package callback

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"agent-telegram/internal/ipc"
)

const (
	verifyTimeout = 5 * time.Second
	maxBodySize   = 1 << 16 // 64 KB
	shutdownGrace = 5 * time.Second
)

// Server is the HTTP API server for callback management.
type Server struct {
	manager *Manager
	port    int
	secret  string
	srv     *http.Server
}

// NewServer creates the HTTP API server.
// If secret is non-empty, all requests must include the X-Secret header.
func NewServer(manager *Manager, port int, secret string) *Server {
	s := &Server{manager: manager, port: port, secret: secret}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /callback/set-callback-url", s.handleSetCallbackURL)
	mux.HandleFunc("GET /callback/get-callback-info", s.handleGetCallbackInfo)
	mux.HandleFunc("POST /callback/subscribe-channel", s.handleSubscribeChannel)
	mux.HandleFunc("GET /callback/subscriptions-list", s.handleSubscriptionsList)
	mux.HandleFunc("POST /callback/unsubscribe", s.handleUnsubscribe)

	var handler http.Handler = mux
	if secret != "" {
		handler = s.authMiddleware(mux)
	}

	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: ipc.ClientTimeout(),
	}
	return s
}

// authMiddleware enforces X-Secret header authentication.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Secret")), []byte(s.secret)) != 1 {
			writeError(w, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Start starts the HTTP server. Blocks until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	lc := &net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", s.srv.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.srv.Addr, err)
	}

	slog.Info("callback API listening", "addr", s.srv.Addr)

	errCh := make(chan error, 1)
	go func() {
		if err := s.srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		return err
	}

	shutCtx, shutCancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer shutCancel()

	return s.srv.Shutdown(shutCtx) //nolint:contextcheck // parent ctx already cancelled
}

// --- Handlers ---

func (s *Server) handleSetCallbackURL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CallbackURL      string `json:"callbackUrl"`
		CallbackURLSnake string `json:"callback_url"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeError(w, "invalid request body")
		return
	}
	callbackURL := firstNonEmpty(req.CallbackURL, req.CallbackURLSnake)
	if callbackURL == "" {
		writeError(w, "callback_url is required")
		return
	}
	if !strings.HasPrefix(callbackURL, "http://") && !strings.HasPrefix(callbackURL, "https://") {
		writeError(w, "callback_url is not valid")
		return
	}

	verifyCode, err := s.manager.SetCallbackURL(callbackURL)
	if err != nil {
		slog.Error("callback: set URL failed", "error", err)
		writeError(w, "internal error")
		return
	}

	// Verify the URL by sending a POST and expecting verifyCode in response
	if err := s.verifyURL(r.Context(), callbackURL, verifyCode); err != nil {
		slog.Warn("callback: URL verification failed",
			"url", callbackURL, "error", err,
		)
		writeJSON(w, http.StatusOK, map[string]any{
			"status":      "error",
			"error":       fmt.Sprintf("verification failed: %v", err),
			"verify_code": verifyCode,
		})
		return
	}

	// Check URL hasn't changed between verify and confirm (race protection)
	cur := s.manager.store.Get()
	if cur.CallbackURL != callbackURL {
		writeError(w, "callback URL changed during verification")
		return
	}

	//nolint:contextcheck // sender goroutine uses app-level context, not HTTP request context
	if err := s.manager.ConfirmVerification(); err != nil {
		writeError(w, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *Server) handleGetCallbackInfo(w http.ResponseWriter, _ *http.Request) {
	state := s.manager.store.Get()
	if state.CallbackURL == "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "error",
			"error":  "Callback URL not set yet",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"response": map[string]any{
			"url":                  state.CallbackURL,
			"verified":             state.Verified,
			"pending_update_count": s.manager.PendingCount(),
			"last_error_date":      state.LastErrorDate,
			"last_error_message":   state.LastErrorMessage,
		},
	})
}

//nolint:funlen // Handler validates peer, event types, optional subscription edit — inherently verbose
func (s *Server) handleSubscribeChannel(w http.ResponseWriter, r *http.Request) {
	state := s.manager.store.Get()
	if state.CallbackURL == "" || !state.Verified {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "error",
			"error":  "callback_url not set yet",
		})
		return
	}

	var req struct {
		SubscriptionID      *int64 `json:"subscriptionId"`
		SubscriptionIDSnake *int64 `json:"subscription_id"`
		ChannelID           string `json:"channelId"`
		ChannelIDSnake      string `json:"channel_id"`
		EventTypes          string `json:"eventTypes"`  // comma-separated
		EventTypesSnake     string `json:"event_types"` // comma-separated
	}
	if err := decodeBody(r, &req); err != nil {
		writeError(w, "invalid request body")
		return
	}
	channelID := firstNonEmpty(req.ChannelID, req.ChannelIDSnake)
	eventTypesRaw := firstNonEmpty(req.EventTypes, req.EventTypesSnake)
	if channelID == "" {
		writeError(w, "channel_id is required")
		return
	}
	if eventTypesRaw == "" {
		writeError(w, "event_types is required")
		return
	}

	eventTypes := splitTrimmed(eventTypesRaw, ",")
	for _, et := range eventTypes {
		if _, ok := eventTypeToUpdateType[et]; !ok {
			writeError(w, fmt.Sprintf("unknown event type: %s", et))
			return
		}
	}

	// Resolve and normalize channelId via Telegram API (validates it exists)
	resolvedChannelID, err := s.manager.ResolveChannelID(r.Context(), channelID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "error",
			"error":  fmt.Sprintf("channel not found: %s", channelID),
		})
		return
	}

	sub := Subscription{
		Type:       "channel",
		ChannelID:  resolvedChannelID,
		EventTypes: eventTypes,
	}

	var id int64
	subscriptionID := req.SubscriptionID
	if subscriptionID == nil {
		subscriptionID = req.SubscriptionIDSnake
	}
	if subscriptionID != nil {
		// Atomic edit: remove old + add new in one lock
		id, err = s.manager.store.EditSubscription(*subscriptionID, sub)
	} else {
		id, err = s.manager.store.AddSubscription(sub)
	}
	if err != nil {
		slog.Error("callback: subscription failed", "error", err)
		writeError(w, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"response": map[string]any{"subscription_id": id},
	})
}

func (s *Server) handleSubscriptionsList(w http.ResponseWriter, _ *http.Request) {
	state := s.manager.store.Get()
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"response": map[string]any{
			"total_count":   len(state.Subscriptions),
			"subscriptions": state.Subscriptions,
		},
	})
}

func (s *Server) handleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SubscriptionID      int64 `json:"subscriptionId"`
		SubscriptionIDSnake int64 `json:"subscription_id"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeError(w, "invalid request body")
		return
	}
	subscriptionID := req.SubscriptionID
	if subscriptionID == 0 {
		subscriptionID = req.SubscriptionIDSnake
	}
	if subscriptionID == 0 {
		writeError(w, "subscription_id is required")
		return
	}
	if err := s.manager.store.RemoveSubscription(subscriptionID); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "error",
			"error":  "subscription not found",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

// --- Verification ---

// verifyURL sends a POST to the URL and expects the verifyCode in the response body.
func (s *Server) verifyURL(ctx context.Context, url, verifyCode string) error {
	payload, err := json.Marshal(map[string]string{"verifyCode": verifyCode})
	if err != nil {
		return fmt.Errorf("marshal verify payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, verifyTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: verifyTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if !strings.Contains(string(body), verifyCode) {
		return fmt.Errorf("verify code not found in response")
	}
	return nil
}

// --- Helpers ---

func decodeBody(r *http.Request, v any) error {
	return json.NewDecoder(io.LimitReader(r.Body, maxBodySize)).Decode(v)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	//nolint:errchkjson // best-effort response write
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "error",
		"error":  msg,
	})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func splitTrimmed(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}
