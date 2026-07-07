package ipc

import (
	"os"
	"time"
)

const (
	// EnvRPCTimeout controls the server-side request timeout and the client
	// socket deadline. It accepts Go duration strings such as "45s" or "2m".
	EnvRPCTimeout = "AGENT_TELEGRAM_RPC_TIMEOUT"

	defaultRequestTimeout = 30 * time.Second
	clientTimeoutGrace    = 5 * time.Second
)

// RequestTimeout returns the timeout used by RPC handlers.
func RequestTimeout() time.Duration {
	if raw := os.Getenv(EnvRPCTimeout); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d > 0 {
			return d
		}
	}
	return defaultRequestTimeout
}

// ClientTimeout returns the client socket deadline. It intentionally adds a
// small grace period so the server can return a typed timeout response.
func ClientTimeout() time.Duration {
	return RequestTimeout() + clientTimeoutGrace
}
