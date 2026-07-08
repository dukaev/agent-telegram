package ipc

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"agent-telegram/internal/observability"
	"agent-telegram/internal/operations"
	"agent-telegram/internal/skills"
)

const (
	maxAPIBodySize   = 1 << 20 // 1 MB
	httpShutdownWait = 5 * time.Second
)

// HTTPServer serves registered IPC handlers over HTTP.
// It implements MethodRegistrar so RegisterHandlers works identically to SocketServer.
type HTTPServer struct {
	methods map[string]Handler
	mu      sync.RWMutex
	secret  string
	cors    string
	srv     *http.Server
	policy  PolicyChecker
}

// NewHTTPServer creates a new HTTP API server.
func NewHTTPServer(port int, secret, cors string) *HTTPServer {
	s := &HTTPServer{
		methods: make(map[string]Handler),
		secret:  secret,
		cors:    cors,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /methods", s.handleMethods)
	mux.HandleFunc("GET /manifest", s.handleManifest)
	mux.HandleFunc("GET /openapi.json", s.handleOpenAPI)
	mux.HandleFunc("POST /rpc/{method}", s.handleRPC)

	var handler http.Handler = mux
	if cors != "" {
		handler = s.corsMiddleware(handler)
	}
	if secret != "" {
		handler = s.authMiddleware(handler)
	}
	handler = s.loggingMiddleware(handler)

	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: ClientTimeout(),
	}
	return s
}

// Register implements MethodRegistrar.
func (s *HTTPServer) Register(name string, handler Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.methods[name] = handler
}

// SetPolicyChecker sets the local policy checker used for HTTP calls.
func (s *HTTPServer) SetPolicyChecker(policy PolicyChecker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policy = policy
}

// Start starts the HTTP server. Blocks until ctx is cancelled.
func (s *HTTPServer) Start(ctx context.Context) error {
	lc := &net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", s.srv.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.srv.Addr, err)
	}

	slog.Info("HTTP API listening", "addr", s.srv.Addr)

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

	shutCtx, shutCancel := context.WithTimeout(context.Background(), httpShutdownWait)
	defer shutCancel()

	return s.srv.Shutdown(shutCtx) //nolint:contextcheck // parent ctx already cancelled
}

// --- Handlers ---

func (s *HTTPServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSONResponse(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *HTTPServer) handleMethods(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	names := make([]string, 0, len(s.methods))
	for name := range s.methods {
		names = append(names, name)
	}
	s.mu.RUnlock()

	sort.Strings(names)
	writeJSONResponse(w, http.StatusOK, map[string]any{"ok": true, "methods": names})
}

func (s *HTTPServer) handleManifest(w http.ResponseWriter, _ *http.Request) {
	writeJSONResponse(w, http.StatusOK, map[string]any{
		"ok":         true,
		"operations": operations.Manifest(),
		"errorTypes": ErrorTypesManifest(),
		"skills":     skills.Manifest(),
	})
}

func (s *HTTPServer) handleOpenAPI(w http.ResponseWriter, _ *http.Request) {
	writeJSONResponse(w, http.StatusOK, operations.OpenAPI("agent-telegram API", "dev"))
}

func (s *HTTPServer) handleRPC(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	method := r.PathValue("method")
	traceID := strings.TrimSpace(r.Header.Get("X-Trace-Id"))
	if traceID == "" {
		traceID = observability.NewTraceID()
	}
	runID := strings.TrimSpace(r.Header.Get("X-Run-Id"))
	if runID != "" {
		runID = observability.SanitizeRunID(runID)
	}
	r.Header.Set("X-Trace-Id", traceID)
	w.Header().Set("X-Trace-Id", traceID)
	if runID != "" {
		r.Header.Set("X-Run-Id", runID)
		w.Header().Set("X-Run-Id", runID)
	}

	s.mu.RLock()
	handler, ok := s.methods[method]
	s.mu.RUnlock()

	if !ok {
		rpcErr := NewTypedError(ErrCodeMethodNotFound, ErrorTypeMethodNotFound, "method not found", nil)
		s.writeHTTPAudit(runID, traceID, method, nil, nil, rpcErr, time.Since(start), false)
		writeJSONResponse(w, http.StatusNotFound, map[string]any{
			"ok":      false,
			"runId":   runID,
			"traceId": traceID,
			"error":   errorResponse(rpcErr),
		})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxAPIBodySize+1))
	if err != nil {
		rpcErr := NewTypedError(ErrCodeParseError, ErrorTypeValidation, "failed to read request body", nil)
		s.writeHTTPAudit(runID, traceID, method, nil, nil, rpcErr, time.Since(start), false)
		writeJSONResponse(w, http.StatusBadRequest, map[string]any{
			"ok":      false,
			"runId":   runID,
			"traceId": traceID,
			"error":   errorResponse(rpcErr),
		})
		return
	}
	if len(body) > maxAPIBodySize {
		rpcErr := NewTypedError(ErrCodeInvalidRequest, ErrorTypeValidation, "request body too large", nil)
		s.writeHTTPAudit(runID, traceID, method, nil, nil, rpcErr, time.Since(start), false)
		writeJSONResponse(w, http.StatusRequestEntityTooLarge, map[string]any{
			"ok":      false,
			"runId":   runID,
			"traceId": traceID,
			"error":   errorResponse(rpcErr),
		})
		return
	}

	params, dryRun, validateOnly, err := parseRPCBody(body)
	if err != nil {
		rpcErr := NewTypedError(ErrCodeInvalidParams, ErrorTypeValidation, err.Error(), nil)
		s.writeHTTPAudit(runID, traceID, method, nil, nil, rpcErr, time.Since(start), false)
		writeJSONResponse(w, errorToHTTPStatus(rpcErr), map[string]any{
			"ok":      false,
			"runId":   runID,
			"traceId": traceID,
			"error":   errorResponse(rpcErr),
		})
		return
	}
	if isTruthy(r.URL.Query().Get("dryRun")) {
		dryRun = true
	}
	if isTruthy(r.URL.Query().Get("validateOnly")) {
		validateOnly = true
	}
	s.mu.RLock()
	policyChecker := s.policy
	s.mu.RUnlock()
	if policyChecker != nil {
		if err := policyChecker.Check(r.Context(), method, params); err != nil {
			rpcErr := ErrorObjectFromError(err)
			if rpcErr == nil {
				rpcErr = NewTypedError(ErrCodeForbidden, ErrorTypeForbidden, err.Error(), nil)
			}
			s.writeHTTPAudit(runID, traceID, method, params, nil, rpcErr, time.Since(start), dryRun || validateOnly)
			writeJSONResponse(w, errorToHTTPStatus(rpcErr), map[string]any{
				"ok":      false,
				"runId":   runID,
				"traceId": traceID,
				"error":   errorResponse(rpcErr),
			})
			return
		}
	}
	if dryRun || validateOnly {
		if err := operations.ValidateParams(method, params); err != nil {
			rpcErr := NewTypedError(ErrCodeInvalidParams, ErrorTypeValidation, err.Error(), nil)
			s.writeHTTPAudit(runID, traceID, method, params, nil, rpcErr, time.Since(start), true)
			writeJSONResponse(w, errorToHTTPStatus(rpcErr), map[string]any{
				"ok":      false,
				"runId":   runID,
				"traceId": traceID,
				"error":   errorResponse(rpcErr),
			})
			return
		}
		op, _ := operations.Get(method)
		s.writeHTTPAudit(
			runID,
			traceID,
			method,
			params,
			map[string]any{"dryRun": dryRun, "validateOnly": validateOnly},
			nil,
			time.Since(start),
			true,
		)
		writeJSONResponse(w, http.StatusOK, map[string]any{
			"ok":           true,
			"runId":        runID,
			"traceId":      traceID,
			"dryRun":       dryRun,
			"validateOnly": validateOnly,
			"method":       method,
			"params":       params,
			"safety":       op.Safety,
			"idempotent":   op.Idempotent,
		})
		return
	}

	result, rpcErr := handler(params)
	if rpcErr != nil {
		status := errorToHTTPStatus(rpcErr)
		s.writeHTTPAudit(runID, traceID, method, params, nil, rpcErr, time.Since(start), false)
		writeJSONResponse(w, status, map[string]any{
			"ok":      false,
			"runId":   runID,
			"traceId": traceID,
			"error":   errorResponse(rpcErr),
		})
		return
	}

	s.writeHTTPAudit(runID, traceID, method, params, result, nil, time.Since(start), false)
	writeJSONResponse(w, http.StatusOK, map[string]any{"ok": true, "runId": runID, "traceId": traceID, "result": result})
}

// --- Middleware ---

func (s *HTTPServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health endpoint
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(s.secret)) != 1 {
			writeJSONResponse(w, http.StatusUnauthorized, map[string]any{
				"ok":    false,
				"error": map[string]any{"code": ErrCodeNotAuthorized, "message": "unauthorized"},
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := allowedCORSOrigin(s.cors, r.Header.Get("Origin"))
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func allowedCORSOrigin(allowed, requestOrigin string) string {
	if allowed == "*" {
		return "*"
	}
	if requestOrigin == "" {
		return ""
	}
	for _, item := range strings.Split(allowed, ",") {
		if strings.TrimSpace(item) == requestOrigin {
			return requestOrigin
		}
	}
	return ""
}

func (s *HTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		slog.Info("http request",
			"run_id", r.Header.Get("X-Run-Id"),
			"trace_id", r.Header.Get("X-Trace-Id"),
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// --- Helpers ---

func errorToHTTPStatus(err *ErrorObject) int {
	switch err.Code {
	case ErrCodeParseError, ErrCodeInvalidRequest, ErrCodeInvalidParams:
		return http.StatusBadRequest
	case ErrCodeMethodNotFound:
		return http.StatusNotFound
	case ErrCodeNotAuthorized:
		return http.StatusForbidden
	case ErrCodeNotInitialized:
		return http.StatusServiceUnavailable
	case ErrCodeTimeout:
		return http.StatusGatewayTimeout
	case ErrCodePeerNotFound:
		return http.StatusNotFound
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeFloodWait:
		return http.StatusTooManyRequests
	case ErrCodePolicyDenied:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func errorResponse(err *ErrorObject) map[string]any {
	out := map[string]any{
		"code":    err.Code,
		"message": err.Message,
	}
	if err.Data != nil {
		out["data"] = err.Data
	}
	return out
}

func (s *HTTPServer) writeHTTPAudit(
	runID string,
	traceID, method string,
	params, result any,
	rpcErr *ErrorObject,
	duration time.Duration,
	dryRun bool,
) {
	event := observability.AuditEvent{
		Time:       time.Now().UTC(),
		RunID:      runID,
		TraceID:    traceID,
		Surface:    "http",
		Method:     method,
		Safety:     operationSafety(method),
		DryRun:     dryRun,
		Status:     "ok",
		DurationMs: duration.Milliseconds(),
		Params:     params,
	}
	if rpcErr != nil {
		event.Status = "error"
		event.ErrorCode = rpcErr.Code
		event.ErrorType = errorType(rpcErr)
		event.Error = rpcErr.Message
	} else {
		event.ResultSummary = observability.SummarizeResult(result)
	}
	if err := observability.WriteAudit("", event); err != nil {
		slog.Warn("http audit write failed", "trace_id", traceID, "error", err)
	}
}

func operationSafety(method string) string {
	if op, ok := operations.Get(method); ok {
		return op.Safety
	}
	return ""
}

func errorType(err *ErrorObject) string {
	if err == nil || err.Data == nil {
		return ""
	}
	if m, ok := err.Data.(map[string]any); ok {
		if value, ok := m["type"].(string); ok {
			return value
		}
	}
	return ""
}

func parseRPCBody(body []byte) (params json.RawMessage, dryRun, validateOnly bool, err error) {
	if len(body) == 0 {
		return json.RawMessage(`{}`), false, false, nil
	}

	var envelope struct {
		Params       json.RawMessage `json:"params"`
		DryRun       bool            `json:"dryRun"`
		ValidateOnly bool            `json:"validateOnly"`
	}
	if json.Unmarshal(body, &envelope) == nil && len(envelope.Params) > 0 {
		return envelope.Params, envelope.DryRun, envelope.ValidateOnly, nil
	}
	if !json.Valid(body) {
		return nil, false, false, fmt.Errorf("invalid JSON body")
	}
	return json.RawMessage(body), false, false, nil
}

func isTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "t", "true", "yes", "y":
		return true
	default:
		return false
	}
}

// ErrorTypesManifest returns stable machine-readable error metadata for agents.
func ErrorTypesManifest() []map[string]any {
	return []map[string]any{
		{"type": ErrorTypeServerNotRunning, "code": ErrCodeServerNotRunning, "retryable": true},
		{"type": ErrorTypeMethodNotFound, "code": ErrCodeMethodNotFound, "retryable": false},
		{"type": ErrorTypeNotAuthorized, "code": ErrCodeNotAuthorized, "retryable": false},
		{"type": ErrorTypeNotInitialized, "code": ErrCodeNotInitialized, "retryable": true},
		{"type": ErrorTypeTimeout, "code": ErrCodeTimeout, "retryable": true},
		{"type": ErrorTypePeerNotFound, "code": ErrCodePeerNotFound, "retryable": false},
		{"type": ErrorTypeForbidden, "code": ErrCodeForbidden, "retryable": false},
		{"type": ErrorTypeFloodWait, "code": ErrCodeFloodWait, "retryable": true},
		{"type": ErrorTypePolicyDenied, "code": ErrCodePolicyDenied, "retryable": false},
		{"type": ErrorTypeValidation, "code": ErrCodeInvalidParams, "retryable": false},
		{"type": ErrorTypeInternal, "code": -32000, "retryable": false},
	}
}

func writeJSONResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	//nolint:errchkjson // best-effort response write
	_ = json.NewEncoder(w).Encode(v)
}

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
