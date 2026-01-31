// Package ipc provides inter-process communication via JSON-RPC.
package ipc

import "encoding/json"

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *ErrorObject  `json:"error,omitempty"`
	ID      interface{}   `json:"id"`
}

// ErrorObject represents a JSON-RPC 2.0 error.
type ErrorObject struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Handler handles JSON-RPC method calls.
type Handler func(params json.RawMessage) (interface{}, *ErrorObject)

// MethodRegistrar is an interface for registering RPC methods.
type MethodRegistrar interface {
	Register(name string, handler Handler)
}
