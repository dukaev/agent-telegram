// Package ipc provides generic handler for Telegram IPC.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	baseipc "agent-telegram/internal/ipc"
	"agent-telegram/internal/strictjson"
	"agent-telegram/telegram/types"
)

// HandlerFunc is the type for IPC handler functions.
// Accepts a context for cancellation and timeout propagation.
type HandlerFunc = func(ctx context.Context, params json.RawMessage) (any, error)

// Handler returns a generic JSON-RPC handler for the given params type.
// Automatically runs struct-tag validation (validate:"required") before calling Validate().
func Handler[T any, R any](
	callFn func(context.Context, T) (R, error),
	methodName string,
) HandlerFunc {
	return func(ctx context.Context, params json.RawMessage) (any, error) {
		var p T
		if len(params) > 0 {
			if err := strictjson.Decode(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		// Auto struct-tag validation (validate:"required", embedded Validate())
		if err := types.ValidateStruct(&p); err != nil {
			return nil, err
		}
		// Custom validation is optional; simple parameter structs only need tags.
		if validator, ok := any(p).(interface{ Validate() error }); ok {
			if err := validator.Validate(); err != nil {
				return nil, err
			}
		}

		result, err := callFn(ctx, p)
		if err != nil {
			return nil, fmt.Errorf("failed to %s: %w", methodName, err)
		}

		return result, nil
	}
}

// ValidateFileParams enforces the HTTP server-side file allowlist before a
// Telegram handler can open a path supplied by a remote caller.
func ValidateFileParams(ctx context.Context, params json.RawMessage) error {
	if baseipc.SurfaceFromContext(ctx) != baseipc.SurfaceHTTP || len(params) == 0 {
		return nil
	}
	var payload struct {
		File string `json:"file"`
	}
	if err := json.Unmarshal(params, &payload); err != nil || payload.File == "" {
		return nil
	}
	roots := baseipc.FileRootsFromContext(ctx)
	if len(roots) == 0 {
		return fmt.Errorf("server-side file paths are disabled over HTTP")
	}
	file, err := filepath.EvalSymlinks(payload.File)
	if err != nil {
		return fmt.Errorf("resolve file path: %w", err)
	}
	file, err = filepath.Abs(file)
	if err != nil {
		return fmt.Errorf("resolve absolute file path: %w", err)
	}
	for _, root := range roots {
		resolvedRoot, rootErr := filepath.EvalSymlinks(root)
		if rootErr != nil {
			continue
		}
		resolvedRoot, rootErr = filepath.Abs(resolvedRoot)
		if rootErr != nil {
			continue
		}
		rel, relErr := filepath.Rel(resolvedRoot, file)
		if relErr == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil
		}
	}
	return fmt.Errorf("file path is outside configured HTTP file roots")
}

// FileHandler returns a handler that validates file existence before calling the method.
func FileHandler[T any, R any](
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
