package cliutil

import (
	"encoding/json"
	"fmt"
	"os"

	"agent-telegram/internal/observability"
)

// FormatSuccess formats a success message with common fields.
// Output goes to stderr so stdout remains clean for data.
func FormatSuccess(result any, action string) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "%s succeeded!\n", action)
		return
	}

	fmt.Fprintf(os.Stderr, "%s sent successfully!\n", action)
	if id, ok := r["id"].(float64); ok {
		fmt.Fprintf(os.Stderr, "  ID: %d\n", int64(id))
	}
	if peer, ok := r["peer"].(string); ok {
		fmt.Fprintf(os.Stderr, "  Peer: %s\n", peer)
	}
}

const maxLogSize = 1024

// truncateAny returns a truncated JSON string representation of a value.
func truncateAny(v any) string {
	data, err := json.Marshal(observability.RedactAny(v))
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	s := string(data)
	if len(s) > maxLogSize {
		return s[:maxLogSize] + "..."
	}
	return s
}

// ExtractString safely extracts a string from a map.
func ExtractString(m map[string]any, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// ExtractFloat64 safely extracts a float64 from a map (handles float64 and json.Number).
func ExtractFloat64(m map[string]any, key string) float64 {
	switch v := m[key].(type) {
	case float64:
		return v
	case json.Number:
		f, _ := v.Float64()
		return f
	}
	return 0
}

// ExtractInt64 safely extracts an int64 from a map (handles int64, float64, and json.Number).
func ExtractInt64(m map[string]any, key string) int64 {
	switch v := m[key].(type) {
	case int64:
		return v
	case json.Number:
		n, _ := v.Int64()
		return n
	case float64:
		return int64(v)
	}
	return 0
}

// ToMap converts any value to a map[string]any safely.
func ToMap(result any) (map[string]any, bool) {
	m, ok := result.(map[string]any)
	return m, ok
}
