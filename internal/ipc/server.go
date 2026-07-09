// Package ipc provides inter-process communication via JSON-RPC.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"agent-telegram/internal/observability"
	"agent-telegram/internal/operations"
)

// Error codes for JSON-RPC errors.
const (
	// Standard JSON-RPC error codes.
	ErrCodeParseError     = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternalError  = -32603

	// Custom error codes.
	ErrCodeServerNotRunning = -32001
	ErrCodeNotAuthorized    = -32002
	ErrCodeNotInitialized   = -32003
)

var (
	// ErrParseError is returned when JSON parsing fails.
	ErrParseError = &ErrorObject{Code: ErrCodeParseError, Message: "Parse error"}
	// ErrInvalidRequest is returned when the request is invalid.
	ErrInvalidRequest = &ErrorObject{Code: ErrCodeInvalidRequest, Message: "Invalid Request"}
	// ErrMethodNotFound is returned when the method is not found.
	ErrMethodNotFound = NewTypedError(ErrCodeMethodNotFound, ErrorTypeMethodNotFound, "Method not found", nil)
	// ErrInvalidParams is returned when the parameters are invalid.
	ErrInvalidParams = &ErrorObject{Code: ErrCodeInvalidParams, Message: "Invalid params"}
	// ErrInternalError is returned when an internal error occurs.
	ErrInternalError = &ErrorObject{Code: ErrCodeInternalError, Message: "Internal error"}

	// ErrServerNotRunning is returned when the server is not running.
	ErrServerNotRunning = NewTypedError(
		ErrCodeServerNotRunning,
		ErrorTypeServerNotRunning,
		"Server is not running",
		nil,
	)
	// ErrNotAuthorized is returned when the user is not authorized.
	ErrNotAuthorized = NewTypedError(
		ErrCodeNotAuthorized,
		ErrorTypeNotAuthorized,
		"Not authorized. Run: agent-telegram auth web",
		nil,
	)
	// ErrNotInitialized is returned when the client is not initialized.
	ErrNotInitialized = NewTypedError(
		ErrCodeNotInitialized,
		ErrorTypeNotInitialized,
		"Client not initialized (server may still be starting)",
		nil,
	)
)

// Server represents a JSON-RPC server.
type Server struct {
	methods map[string]Handler
	policy  PolicyChecker
	mu      sync.RWMutex
}

// NewServer creates a new JSON-RPC server.
func NewServer() *Server {
	return &Server{
		methods: make(map[string]Handler),
	}
}

// Register registers a method handler.
func (s *Server) Register(name string, handler Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.methods[name] = handler
}

// SetPolicyChecker sets the local policy checker used before method execution.
func (s *Server) SetPolicyChecker(policy PolicyChecker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policy = policy
}

// Serve starts the JSON-RPC server on the given io.ReadWriteCloser.
func (s *Server) Serve(ctx context.Context, rwc io.ReadWriteCloser) error {
	defer func() { _ = rwc.Close() }()
	decoder := json.NewDecoder(rwc)
	encoder := json.NewEncoder(rwc)

	for {
		var req Request
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			slog.Warn("ipc: failed to decode request", "error", err)
			s.sendError(encoder, nil, ErrParseError)
			return nil
		}

		resp := s.handleRequest(ctx, &req)
		if req.ID == nil {
			continue
		}

		if err := encoder.Encode(resp); err != nil {
			return fmt.Errorf("encode response: %w", err)
		}
	}
}

// ServeStdinStdout serves JSON-RPC over stdin/stdout.
func (s *Server) ServeStdinStdout(ctx context.Context) error {
	return s.Serve(ctx, &readWriteCloser{
		Reader:    os.Stdin,
		Writer:    os.Stdout,
		CloseFunc: func() error { return nil },
	})
}

func (s *Server) handleRequest(ctx context.Context, req *Request) (resp *Response) {
	start := time.Now()
	ctx = WithConfirmation(ctx, req.Confirm)

	defer func() {
		if r := recover(); r != nil {
			slog.Error("ipc: handler panic", "method", req.Method, "panic", r)
			resp = &Response{
				JSONRPC: "2.0",
				Error: &ErrorObject{
					Code:    -32603,
					Message: fmt.Sprintf("Handler panic: %v", r),
				},
				ID:      req.ID,
				RunID:   req.RunID,
				TraceID: req.TraceID,
			}
		}
		s.logRequest(req, resp, time.Since(start))
	}()
	if req.JSONRPC != "2.0" || req.Method == "" {
		return &Response{
			JSONRPC: "2.0",
			Error:   ErrInvalidRequest,
			ID:      req.ID,
			RunID:   req.RunID,
			TraceID: req.TraceID,
		}
	}

	s.mu.RLock()
	handler, ok := s.methods[req.Method]
	policyChecker := s.policy
	s.mu.RUnlock()

	if !ok {
		return &Response{
			JSONRPC: "2.0",
			Error:   ErrMethodNotFound,
			ID:      req.ID,
			RunID:   req.RunID,
			TraceID: req.TraceID,
		}
	}
	if operations.HasSchema(req.Method) {
		if err := operations.ValidateParams(req.Method, req.Params); err != nil {
			return &Response{
				JSONRPC: "2.0",
				Error:   NewTypedError(ErrCodeInvalidParams, ErrorTypeValidation, err.Error(), nil),
				ID:      req.ID,
				RunID:   req.RunID,
				TraceID: req.TraceID,
			}
		}
	}

	if policyChecker != nil {
		ctx, cancel := context.WithTimeout(ctx, RequestTimeout())
		defer cancel()
		if err := policyChecker.Check(ctx, req.Method, req.Params); err != nil {
			rpcErr := ErrorObjectFromError(err)
			if rpcErr == nil {
				rpcErr = NewTypedError(ErrCodeForbidden, ErrorTypeForbidden, err.Error(), nil)
			}
			return &Response{
				JSONRPC: "2.0",
				Error:   rpcErr,
				ID:      req.ID,
				RunID:   req.RunID,
				TraceID: req.TraceID,
			}
		}
	}

	result, err := handler(ctx, req.Params)
	if err != nil {
		return &Response{
			JSONRPC: "2.0",
			Error:   err,
			ID:      req.ID,
			RunID:   req.RunID,
			TraceID: req.TraceID,
		}
	}

	return &Response{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
		RunID:   req.RunID,
		TraceID: req.TraceID,
	}
}

// logRequest logs an IPC request and its response.
func (s *Server) logRequest(req *Request, resp *Response, duration time.Duration) {
	params := truncateJSON(req.Params, maxLogSize)
	if resp.Error != nil {
		slog.Info("ipc: request",
			"run_id", req.RunID,
			"trace_id", req.TraceID,
			"method", req.Method,
			"params", params,
			"duration_ms", duration.Milliseconds(),
			"error_code", resp.Error.Code,
			"error", resp.Error.Message,
		)
	} else {
		resultJSON, err := json.Marshal(resp.Result)
		if err != nil {
			slog.Debug("ipc: failed to marshal result for logging", "error", err)
		}
		slog.Info("ipc: request",
			"run_id", req.RunID,
			"trace_id", req.TraceID,
			"method", req.Method,
			"params", params,
			"duration_ms", duration.Milliseconds(),
			"result_size", len(resultJSON),
		)
	}
}

const maxLogSize = 1024

// truncateJSON returns a truncated string representation of JSON data.
func truncateJSON(data json.RawMessage, maxLen int) string {
	if len(data) == 0 {
		return "{}"
	}
	var value any
	if err := json.Unmarshal(data, &value); err == nil {
		if redacted, err := json.Marshal(observability.RedactAny(value)); err == nil {
			data = redacted
		}
	}
	s := string(data)
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func (s *Server) sendError(encoder *json.Encoder, id any, err *ErrorObject) {
	resp := &Response{
		JSONRPC: "2.0",
		Error:   err,
		ID:      id,
	}
	if encErr := encoder.Encode(resp); encErr != nil {
		// Last resort error logging
		slog.Error("failed to encode error response", "error", encErr)
	}
}

// readWriteCloser combines io.Reader, io.Writer, and io.Closer.
type readWriteCloser struct {
	io.Reader
	io.Writer
	CloseFunc func() error
}

func (rwc *readWriteCloser) Close() error {
	if rwc.CloseFunc != nil {
		return rwc.CloseFunc()
	}
	return nil
}
