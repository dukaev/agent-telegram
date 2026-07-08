package ipc

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"agent-telegram/internal/observability"
	"agent-telegram/internal/operations"
)

type httpRPCRequest struct {
	start   time.Time
	method  string
	runID   string
	traceID string
}

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

func newHTTPRPCRequest(w http.ResponseWriter, r *http.Request) httpRPCRequest {
	traceID := strings.TrimSpace(r.Header.Get("X-Trace-Id"))
	if traceID == "" {
		traceID = observability.NewTraceID()
	}
	runID := strings.TrimSpace(r.Header.Get("X-Run-Id"))
	if runID != "" {
		runID = observability.SanitizeRunID(runID)
		r.Header.Set("X-Run-Id", runID)
		w.Header().Set("X-Run-Id", runID)
	}
	r.Header.Set("X-Trace-Id", traceID)
	w.Header().Set("X-Trace-Id", traceID)

	return httpRPCRequest{
		start:   time.Now(),
		method:  r.PathValue("method"),
		runID:   runID,
		traceID: traceID,
	}
}

func (s *HTTPServer) lookupHandler(method string) (Handler, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	handler, ok := s.methods[method]
	return handler, ok
}

func readHTTPRPCBody(r *http.Request) ([]byte, *ErrorObject, int) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxAPIBodySize+1))
	if err != nil {
		rpcErr := NewTypedError(ErrCodeParseError, ErrorTypeValidation, "failed to read request body", nil)
		return nil, rpcErr, http.StatusBadRequest
	}
	if len(body) > maxAPIBodySize {
		rpcErr := NewTypedError(ErrCodeInvalidRequest, ErrorTypeValidation, "request body too large", nil)
		return nil, rpcErr, http.StatusRequestEntityTooLarge
	}
	return body, nil, http.StatusOK
}

func parseHTTPRPCParams(r *http.Request, body []byte) (json.RawMessage, bool, bool, *ErrorObject) {
	params, dryRun, validateOnly, err := parseRPCBody(body)
	if err != nil {
		rpcErr := NewTypedError(ErrCodeInvalidParams, ErrorTypeValidation, err.Error(), nil)
		return nil, false, false, rpcErr
	}
	if isTruthy(r.URL.Query().Get("dryRun")) {
		dryRun = true
	}
	if isTruthy(r.URL.Query().Get("validateOnly")) {
		validateOnly = true
	}
	return params, dryRun, validateOnly, nil
}

func (s *HTTPServer) checkHTTPPolicy(ctx context.Context, method string, params json.RawMessage) *ErrorObject {
	s.mu.RLock()
	policyChecker := s.policy
	s.mu.RUnlock()
	if policyChecker == nil {
		return nil
	}

	if err := policyChecker.Check(ctx, method, params); err != nil {
		if rpcErr := ErrorObjectFromError(err); rpcErr != nil {
			return rpcErr
		}
		return NewTypedError(ErrCodeForbidden, ErrorTypeForbidden, err.Error(), nil)
	}
	return nil
}

func (s *HTTPServer) writeRPCError(
	w http.ResponseWriter,
	req httpRPCRequest,
	params any,
	rpcErr *ErrorObject,
	status int,
	auditDryRun bool,
) {
	s.writeHTTPAudit(req.runID, req.traceID, req.method, params, nil, rpcErr, time.Since(req.start), auditDryRun)
	writeJSONResponse(w, status, map[string]any{
		"ok":      false,
		"runId":   req.runID,
		"traceId": req.traceID,
		"error":   errorResponse(rpcErr),
	})
}

func (s *HTTPServer) handleHTTPDryRun(
	w http.ResponseWriter,
	req httpRPCRequest,
	params json.RawMessage,
	dryRun bool,
	validateOnly bool,
) {
	if err := operations.ValidateParams(req.method, params); err != nil {
		rpcErr := NewTypedError(ErrCodeInvalidParams, ErrorTypeValidation, err.Error(), nil)
		s.writeRPCError(w, req, params, rpcErr, errorToHTTPStatus(rpcErr), true)
		return
	}

	op, _ := operations.Get(req.method)
	result := map[string]any{"dryRun": dryRun, "validateOnly": validateOnly}
	s.writeHTTPAudit(req.runID, req.traceID, req.method, params, result, nil, time.Since(req.start), true)
	writeJSONResponse(w, http.StatusOK, map[string]any{
		"ok":           true,
		"runId":        req.runID,
		"traceId":      req.traceID,
		"dryRun":       dryRun,
		"validateOnly": validateOnly,
		"method":       req.method,
		"params":       params,
		"safety":       op.Safety,
		"idempotent":   op.Idempotent,
	})
}

func (s *HTTPServer) writeRPCSuccess(w http.ResponseWriter, req httpRPCRequest, params, result any) {
	s.writeHTTPAudit(req.runID, req.traceID, req.method, params, result, nil, time.Since(req.start), false)
	writeJSONResponse(w, http.StatusOK, map[string]any{
		"ok":      true,
		"runId":   req.runID,
		"traceId": req.traceID,
		"result":  result,
	})
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
