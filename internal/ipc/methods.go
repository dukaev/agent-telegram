// Package ipc provides inter-process communication via JSON-RPC.
package ipc

import (
	"encoding/json"
)

// PingParams represents the parameters for a ping request.
type PingParams struct {
	Message string `json:"message,omitempty"`
}

// PingResult represents the result of a ping request.
type PingResult struct {
	Message string `json:"message"`
	Pong    bool   `json:"pong"`
}

// RegisterPingPong registers ping/pong methods on the server.
func RegisterPingPong(srv MethodRegistrar) {
	srv.Register("ping", func(params json.RawMessage) (interface{}, *ErrorObject) {
		var p PingParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, ErrInvalidParams
			}
		}
		return PingResult{
			Message: "pong",
			Pong:    true,
		}, nil
	})

	srv.Register("echo", func(params json.RawMessage) (interface{}, *ErrorObject) {
		var p PingParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, ErrInvalidParams
			}
		}
		return map[string]interface{}{
			"echo": p.Message,
		}, nil
	})
}
