package cliutil

import (
	"encoding/json"
	"fmt"
	"strings"

	"agent-telegram/internal/observability"
)

// Verbosity controls how much structured output is emitted.
type Verbosity string

const (
	VerbosityMinimal Verbosity = "minimal"
	VerbosityCompact Verbosity = "compact"
	VerbosityFull    Verbosity = "full"
	VerbosityRaw     Verbosity = "raw"
)

// OutputBudgetOptions describes token-saving output constraints.
type OutputBudgetOptions struct {
	Verbosity    Verbosity
	MaxItems     int
	MaxTextChars int
	Include      []string
	Omit         []string
	Summary      bool
}

// ParseVerbosity parses a verbosity flag value.
func ParseVerbosity(value string) Verbosity {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(VerbosityMinimal):
		return VerbosityMinimal
	case string(VerbosityCompact):
		return VerbosityCompact
	case string(VerbosityRaw):
		return VerbosityRaw
	default:
		return VerbosityFull
	}
}

// ApplyOutputBudget applies verbosity, size, include, omit, and summary settings.
func ApplyOutputBudget(result any, opts OutputBudgetOptions) any {
	if opts.Summary {
		return observability.SummarizeResult(result)
	}
	if opts.Verbosity == "" {
		opts.Verbosity = VerbosityFull
	}
	if opts.Verbosity == VerbosityRaw {
		return result
	}

	normalized := normalizeForBudget(result)
	normalized = applyProfile(normalized, opts)
	if len(opts.Include) > 0 {
		normalized = NewFieldSelector(opts.Include).Apply(normalized)
	}
	if len(opts.Omit) > 0 {
		normalized = omitFields(normalized, opts.Omit)
	}
	return normalized
}

func normalizeForBudget(result any) any {
	data, err := json.Marshal(result)
	if err != nil {
		return result
	}
	var out any
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	if err := decoder.Decode(&out); err != nil {
		return result
	}
	return out
}

func applyProfile(value any, opts OutputBudgetOptions) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, child := range v {
			if arr, ok := child.([]any); ok {
				out[key] = budgetArray(key, arr, opts)
			} else if s, ok := child.(string); ok {
				out[key] = budgetString(key, s, opts)
			} else {
				out[key] = applyProfile(child, opts)
			}
		}
		return out
	case []any:
		return budgetArray("", v, opts)
	case string:
		return budgetString("", v, opts)
	default:
		return v
	}
}

func budgetArray(key string, arr []any, opts OutputBudgetOptions) []any {
	limit := opts.MaxItems
	if limit <= 0 {
		switch opts.Verbosity {
		case VerbosityMinimal, VerbosityCompact:
			limit = 20
		default:
			limit = len(arr)
		}
	}
	if limit > len(arr) {
		limit = len(arr)
	}
	if shouldTailArray(key) && limit < len(arr) {
		arr = arr[len(arr)-limit:]
	} else {
		arr = arr[:limit]
	}
	out := make([]any, 0, limit)
	for _, item := range arr {
		if opts.Verbosity == VerbosityMinimal {
			out = append(out, minimalItem(item))
			continue
		}
		out = append(out, applyProfile(item, opts))
	}
	return out
}

func shouldTailArray(key string) bool {
	switch strings.ToLower(key) {
	case "events", "lines", "logs", "audit":
		return true
	default:
		return false
	}
}

func minimalItem(item any) any {
	m, ok := item.(map[string]any)
	if !ok {
		return item
	}
	out := make(map[string]any)
	for _, key := range []string{
		"id", "messageId", "message_id", "peer", "username", "title", "type", "date",
		"from", "chat", "sender", "success", "status", "count", "total",
	} {
		if value, exists := m[key]; exists {
			out[key] = value
		}
	}
	for _, key := range []string{"text", "message", "caption"} {
		if value, ok := m[key].(string); ok && value != "" {
			out[key+"Preview"] = truncateRunes(value, 80)
			break
		}
	}
	if len(out) == 0 {
		return m
	}
	return out
}

func budgetString(key, value string, opts OutputBudgetOptions) string {
	if !isTextLikeKey(key) && key != "" {
		return value
	}
	maxChars := opts.MaxTextChars
	if maxChars <= 0 {
		switch opts.Verbosity {
		case VerbosityMinimal:
			maxChars = 80
		case VerbosityCompact:
			maxChars = 200
		}
	}
	if maxChars <= 0 {
		return value
	}
	return truncateRunes(value, maxChars)
}

func isTextLikeKey(key string) bool {
	switch strings.ToLower(key) {
	case "text", "message", "caption", "description", "bio", "about":
		return true
	default:
		return false
	}
}

func truncateRunes(value string, max int) string {
	if max <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max]) + fmt.Sprintf("... [truncated %d chars]", len(runes)-max)
}

func omitFields(value any, fields []string) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, child := range v {
			if containsField(fields, key) {
				continue
			}
			out[key] = omitNested(child, fields)
		}
		return out
	case []any:
		return omitNested(v, fields)
	default:
		return value
	}
}

func omitNested(value any, fields []string) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, child := range v {
			if containsField(fields, key) {
				continue
			}
			out[key] = omitNested(child, fields)
		}
		for _, field := range fields {
			if strings.Contains(field, ".") {
				deleteNested(out, strings.Split(field, "."))
			}
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = omitNested(item, fields)
		}
		return out
	default:
		return value
	}
}

func containsField(fields []string, key string) bool {
	for _, field := range fields {
		if field == key {
			return true
		}
	}
	return false
}

func deleteNested(m map[string]any, parts []string) {
	if len(parts) == 0 {
		return
	}
	if len(parts) == 1 {
		delete(m, parts[0])
		return
	}
	child, ok := m[parts[0]].(map[string]any)
	if !ok {
		return
	}
	deleteNested(child, parts[1:])
}
