package cliutil

import (
	"strings"
)

// FieldSelector filters map keys to only include specified fields.
type FieldSelector struct {
	Fields []string
}

// NewFieldSelector creates a FieldSelector from a list of field names.
// Returns nil if fields is empty.
func NewFieldSelector(fields []string) *FieldSelector {
	if len(fields) == 0 {
		return nil
	}
	return &FieldSelector{Fields: fields}
}

// Apply filters the result to include only the selected fields.
// For wrapper maps with item arrays, it filters each item's keys while keeping metadata.
// For single maps, it filters keys directly.
// Supports dot notation for nested fields (e.g., "from.id").
func (fs *FieldSelector) Apply(result any) any {
	if fs == nil || len(fs.Fields) == 0 {
		return result
	}

	rMap, ok := result.(map[string]any)
	if !ok {
		return result
	}

	// Check if this is a wrapper with an items array
	for _, key := range knownArrayKeys {
		if arr, ok := rMap[key].([]any); ok {
			filtered := make([]any, 0, len(arr))
			for _, item := range arr {
				if itemMap, ok := item.(map[string]any); ok {
					filtered = append(filtered, fs.filterMap(itemMap))
				} else {
					filtered = append(filtered, item)
				}
			}
			// Reconstruct wrapper with metadata
			out := make(map[string]any)
			for k, v := range rMap {
				if k == key {
					out[k] = filtered
				} else {
					out[k] = v // keep metadata (count, offset, etc.)
				}
			}
			return out
		}
	}

	// Single map — filter keys directly
	return fs.filterMap(rMap)
}

// filterMap returns a new map containing only the selected fields.
func (fs *FieldSelector) filterMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(fs.Fields))
	for _, field := range fs.Fields {
		if strings.Contains(field, ".") {
			// Dot notation: extract nested value
			if val := extractNested(m, field); val != nil {
				out[field] = val
			}
		} else {
			if val, ok := m[field]; ok {
				out[field] = val
			}
		}
	}
	return out
}

// extractNested extracts a value from nested maps using dot notation.
func extractNested(m map[string]any, path string) any {
	parts := strings.SplitN(path, ".", 2)
	val, ok := m[parts[0]]
	if !ok {
		return nil
	}
	if len(parts) == 1 {
		return val
	}
	nested, ok := val.(map[string]any)
	if !ok {
		return nil
	}
	return extractNested(nested, parts[1])
}
