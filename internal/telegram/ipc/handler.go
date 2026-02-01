// Package ipc provides generic handler for Telegram IPC.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// HandlerFunc is the type for IPC handler functions.
type HandlerFunc = func(json.RawMessage) (any, error)

// Params interface for types that can validate themselves.
type Params interface {
	Validate() error
}

// Handler returns a generic JSON-RPC handler for the given params type.
func Handler[T Params, R any](
	callFn func(context.Context, T) (R, error),
	methodName string,
) HandlerFunc {
	return func(params json.RawMessage) (any, error) {
		var p T
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := callFn(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to %s: %w", methodName, err)
		}

		return result, nil
	}
}

// FileHandler returns a handler that validates file existence before calling the method.
func FileHandler[T Params, R any](
	getFile func(T) string,
	callFn func(context.Context, T) (R, error),
	methodName string,
) HandlerFunc {
	return Handler(func(ctx context.Context, p T) (R, error) {
		file := getFile(p)
		if _, err := os.Stat(file); os.IsNotExist(err) {
			var zero R
			return zero, fmt.Errorf("file not found: %s", file)
		}
		return callFn(ctx, p)
	}, methodName)
}
