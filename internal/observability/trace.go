// Package observability provides trace IDs, redaction, and audit journaling.
package observability

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
)

const (
	envTraceID = "AGENT_TELEGRAM_TRACE_ID"
	envRunID   = "AGENT_TELEGRAM_RUN_ID"
)

// NewTraceID returns a stable trace ID for one user-facing operation.
func NewTraceID() string {
	if value := strings.TrimSpace(os.Getenv(envTraceID)); value != "" {
		return sanitizeID(value, "trace-unavailable")
	}
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "trace-unavailable"
	}
	return hex.EncodeToString(buf[:])
}

// NewRunID returns a stable run ID for a multi-command agent task.
func NewRunID() string {
	if value := strings.TrimSpace(os.Getenv(envRunID)); value != "" {
		return SanitizeRunID(value)
	}
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "run-unavailable"
	}
	return "run_" + hex.EncodeToString(buf[:])
}

// SanitizeRunID keeps caller-provided run IDs log-friendly.
func SanitizeRunID(value string) string {
	return sanitizeID(value, "run-unavailable")
}

func sanitizeID(value, fallback string) string {
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
		return fallback
	}
	return b.String()
}
