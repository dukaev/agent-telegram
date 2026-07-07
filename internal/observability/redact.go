package observability

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

const redacted = "[REDACTED]"

// RedactAny returns a copy of value with secrets masked.
func RedactAny(value any) any {
	normalized := normalizeJSON(value)
	return redactValue("", normalized)
}

// SummarizeResult returns a small, safe result summary for audit logs.
func SummarizeResult(value any) map[string]any {
	normalized := normalizeJSON(value)
	summary := map[string]any{"type": fmt.Sprintf("%T", value)}
	if m, ok := normalized.(map[string]any); ok {
		for _, key := range []string{"success", "messageId", "message_id", "id", "count", "total", "limit", "offset"} {
			if val, exists := m[key]; exists {
				summary[key] = val
			}
		}
		for key, value := range m {
			arr, ok := value.([]any)
			if !ok {
				continue
			}
			summary[key+"Count"] = len(arr)
		}
		for _, key := range []string{"messages", "chats", "contacts", "items", "updates", "gifts"} {
			if arr, ok := m[key].([]any); ok {
				summary[key+"Count"] = len(arr)
			}
		}
		if len(summary) == 1 {
			summary["keys"] = mapKeys(m)
		}
		return RedactAny(summary).(map[string]any)
	}
	if arr, ok := normalized.([]any); ok {
		summary["count"] = len(arr)
	}
	return summary
}

func normalizeJSON(value any) any {
	if value == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	var out any
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	if err := decoder.Decode(&out); err != nil {
		return fmt.Sprint(value)
	}
	return out
}

func redactValue(key string, value any) any {
	if isSensitiveKey(key) {
		return redacted
	}
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for childKey, childValue := range v {
			out[childKey] = redactValue(childKey, childValue)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = redactValue(key, item)
		}
		return out
	case string:
		return redactStringForKey(key, v)
	default:
		return v
	}
}

func redactStringForKey(key, value string) string {
	lower := strings.ToLower(key)
	if strings.Contains(lower, "path") && strings.Contains(strings.ToLower(value), "session") {
		return filepath.Base(value)
	}
	if strings.Contains(lower, "phone") {
		return MaskPhone(value)
	}
	if isTextLikeKey(lower) {
		return truncateText(value, 80)
	}
	return value
}

// MaskPhone masks all but the last four digits in a phone-like string.
func MaskPhone(phone string) string {
	digits := make([]rune, 0, len(phone))
	for _, r := range phone {
		if unicode.IsDigit(r) {
			digits = append(digits, r)
		}
	}
	if len(digits) <= 4 {
		return "***"
	}
	return "***" + string(digits[len(digits)-4:])
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "_", ""), "-", ""))
	if normalized == "" {
		return false
	}
	sensitive := []string{
		"token", "secret", "password", "passcode", "code", "otp",
		"phonecodehash", "apphash", "authorization", "session",
	}
	for _, item := range sensitive {
		if strings.Contains(normalized, item) {
			return true
		}
	}
	return false
}

func isTextLikeKey(key string) bool {
	switch key {
	case "text", "message", "caption":
		return true
	default:
		return false
	}
}

func truncateText(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max]) + fmt.Sprintf("... [truncated %d chars]", len(runes)-max)
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
