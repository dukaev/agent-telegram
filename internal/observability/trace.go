// Package observability provides trace IDs, redaction, and audit journaling.
package observability

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
)

const envTraceID = "AGENT_TELEGRAM_TRACE_ID"

// NewTraceID returns a stable trace ID for one user-facing operation.
func NewTraceID() string {
	if value := strings.TrimSpace(os.Getenv(envTraceID)); value != "" {
		return sanitizeTraceID(value)
	}
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "trace-unavailable"
	}
	return hex.EncodeToString(buf[:])
}

// sanitizeTraceID keeps trace IDs log-friendly.
func sanitizeTraceID(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			b.WriteRune(r)
		}
		if b.Len() >= 64 {
			break
		}
	}
	if b.Len() == 0 {
		return "trace-unavailable"
	}
	return b.String()
}
