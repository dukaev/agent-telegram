package observability

import (
	"encoding/json"
	"strings"
)

// RedactionMode controls how much sensitive action data is shown in audit/log output.
type RedactionMode string

const (
	RedactionSafe     RedactionMode = "safe"
	RedactionRedacted RedactionMode = "redacted"
)

// ParseRedactionMode parses a redaction flag value.
func ParseRedactionMode(value string) RedactionMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(RedactionRedacted):
		return RedactionRedacted
	default:
		return RedactionSafe
	}
}

// RedactAuditEventsForDisplay applies display-time redaction to audit events.
func RedactAuditEventsForDisplay(events []AuditEvent, mode RedactionMode) []AuditEvent {
	out := make([]AuditEvent, len(events))
	for i, event := range events {
		out[i] = event
		if mode == RedactionRedacted {
			out[i].Params = RedactAny(event.Params)
			out[i].ResultSummary = redactedSummary(event.ResultSummary)
			continue
		}
		out[i].Params = redactSafeParams(event.Params)
		out[i].ResultSummary = redactedSummary(event.ResultSummary)
	}
	return out
}

// RedactLogLineForDisplay applies best-effort redaction to JSON log lines.
func RedactLogLineForDisplay(line string, mode RedactionMode) string {
	var entry map[string]any
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		return line
	}
	params, ok := entry["params"].(string)
	if ok && params != "" {
		var parsed any
		if err := json.Unmarshal([]byte(params), &parsed); err == nil {
			var redactedParams any
			if mode == RedactionRedacted {
				redactedParams = RedactAny(parsed)
			} else {
				redactedParams = redactSafeParams(parsed)
			}
			if data, err := json.Marshal(redactedParams); err == nil {
				entry["params"] = string(data)
			}
		}
	}
	if data, err := json.Marshal(entry); err == nil {
		return string(data)
	}
	return line
}

func redactSafeParams(value any) any {
	normalized := normalizeJSON(value)
	return redactSafeValue("", normalized)
}

func redactSafeValue(key string, value any) any {
	if isSensitiveKey(key) {
		return redacted
	}
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for childKey, childValue := range v {
			out[childKey] = redactSafeValue(childKey, childValue)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = redactSafeValue(key, item)
		}
		return out
	case string:
		return redactSafeString(key, v)
	case float64, json.Number:
		if isLocationKey(key) || isPersonalKey(key) {
			return "[PERSONAL DATA REDACTED]"
		}
		return v
	default:
		return v
	}
}

func redactSafeString(key, value string) string {
	lower := strings.ToLower(key)
	if strings.Contains(lower, "phone") {
		return MaskPhone(value)
	}
	if isLocationKey(lower) {
		return "[LOCATION REDACTED]"
	}
	if isPersonalKey(lower) {
		return "[PERSONAL DATA REDACTED]"
	}
	if isTextLikeKey(lower) {
		if strings.HasPrefix(strings.TrimSpace(value), "/") {
			return value
		}
		return "[TEXT REDACTED]"
	}
	return redactStringForKey(key, value)
}

func isLocationKey(key string) bool {
	normalized := normalizeKey(key)
	for _, part := range []string{"location", "latitude", "longitude", "city", "address", "geo"} {
		if strings.Contains(normalized, part) {
			return true
		}
	}
	return false
}

func isPersonalKey(key string) bool {
	normalized := normalizeKey(key)
	for _, part := range []string{"birth", "birthday", "dob"} {
		if strings.Contains(normalized, part) {
			return true
		}
	}
	return normalized == "age"
}

func normalizeKey(key string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "_", ""), "-", ""))
}
