// Package ipc provides inter-process communication via JSON-RPC.
package ipc

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

var (
	// ErrParseError is returned when JSON parsing fails.
	ErrParseError = &ErrorObject{Code: -32700, Message: "Parse error"}
	// ErrInvalidRequest is returned when the request is invalid.
	ErrInvalidRequest = &ErrorObject{Code: -32600, Message: "Invalid Request"}
	// ErrMethodNotFound is returned when the method is not found.
	ErrMethodNotFound = &ErrorObject{Code: -32601, Message: "Method not found"}
	// ErrInvalidParams is returned when the parameters are invalid.
	ErrInvalidParams = &ErrorObject{Code: -32602, Message: "Invalid params"}
	// ErrInternalError is returned when an internal error occurs.
	ErrInternalError = &ErrorObject{Code: -32603, Message: "Internal error"}
)

// Server represents a JSON-RPC server.
type Server struct {
	methods map[string]Handler
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

// Serve starts the JSON-RPC server on the given io.ReadWriteCloser.
func (s *Server) Serve(rwc io.ReadWriteCloser) error {
	defer func() { _ = rwc.Close() }()
	decoder := json.NewDecoder(rwc)
	encoder := json.NewEncoder(rwc)

	for {
		var req Request
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			s.sendError(encoder, nil, ErrParseError)
			continue
		}

		// Log incoming request
		slog.Debug("received request", "method", req.Method, "id", req.ID)

		resp := s.handleRequest(&req)

		// Log outgoing response
		if resp.Error != nil {
			slog.Debug("sending error response", "error", resp.Error.Message, "code", resp.Error.Code, "id", resp.ID)
		} else {
			slog.Debug("sending success response", "id", resp.ID)
		}

		if err := encoder.Encode(resp); err != nil {
			return fmt.Errorf("encode response: %w", err)
		}
	}
}

// ServeStdinStdout serves JSON-RPC over stdin/stdout.
func (s *Server) ServeStdinStdout() error {
	return s.Serve(&readWriteCloser{
		Reader:    os.Stdin,
		Writer:    os.Stdout,
		CloseFunc: func() error { return nil },
	})
}

func (s *Server) handleRequest(req *Request) (resp *Response) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("handler panic", "panic", r, "method", req.Method)
			resp = &Response{
				JSONRPC: "2.0",
				Error: &ErrorObject{
					Code:    -32603,
					Message: fmt.Sprintf("Handler panic: %v", r),
				},
				ID: req.ID,
			}
		}
	}()

	s.mu.RLock()
	handler, ok := s.methods[req.Method]
	s.mu.RUnlock()

	if !ok {
		return &Response{
			JSONRPC: "2.0",
			Error:   ErrMethodNotFound,
			ID:      req.ID,
		}
	}

	result, err := handler(req.Params)
	if err != nil {
		return &Response{
			JSONRPC: "2.0",
			Error:   err,
			ID:      req.ID,
		}
	}

	return &Response{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	}
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
