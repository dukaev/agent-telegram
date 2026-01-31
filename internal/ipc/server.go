// Package ipc provides inter-process communication via JSON-RPC.
package ipc

import (
	"encoding/json"
	"fmt"
	"io"
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
		reqJSON, err := json.Marshal(req)
		if err == nil {
			fmt.Printf("→ Received: %s\n", string(reqJSON))
		}

		resp := s.handleRequest(&req)

		// Log outgoing response
		respJSON, err := json.Marshal(resp)
		if err == nil {
			fmt.Printf("← Sending: %s\n", string(respJSON))
		}

		if err := encoder.Encode(resp); err != nil {
			return fmt.Errorf("encode response: %w", err)
		}
	}
}

// ServeStdinStdout serves JSON-RPC over stdin/stdout.
func (s *Server) ServeStdinStdout() error {
	return s.Serve(&readWriteCloser{
		Reader:   os.Stdin,
		Writer:   os.Stdout,
		CloseFunc: func() error { return nil },
	})
}

func (s *Server) handleRequest(req *Request) *Response {
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

func (s *Server) sendError(encoder *json.Encoder, id interface{}, err *ErrorObject) {
	resp := &Response{
		JSONRPC: "2.0",
		Error:   err,
		ID:      id,
	}
	if encErr := encoder.Encode(resp); encErr != nil {
		// Last resort error logging
		fmt.Printf("Failed to encode error response: %v\n", encErr)
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
