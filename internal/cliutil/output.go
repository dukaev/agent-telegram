package cliutil

import (
	"fmt"
	"os"
	"strings"
)

// OutputFormat defines the output format mode.
type OutputFormat string

const (
	// OutputText is the default human-readable format (to stderr).
	OutputText OutputFormat = "text"
	// OutputJSON outputs JSON to stdout.
	OutputJSON OutputFormat = "json"
	// OutputIDs outputs one ID per line to stdout.
	OutputIDs OutputFormat = "ids"
)

// ParseOutputFormat resolves the output format from --output and --text flags.
// Default is JSON. Use --output text or --text for human-readable output.
func ParseOutputFormat(outputFlag string, textFlag bool) OutputFormat {
	if outputFlag != "" {
		switch strings.ToLower(outputFlag) {
		case "json":
			return OutputJSON
		case "ids":
			return OutputIDs
		case "text":
			return OutputText
		default:
			fmt.Fprintf(os.Stderr, "Warning: unknown output format %q, using json\n", outputFlag)
			return OutputJSON
		}
	}
	if textFlag {
		return OutputText
	}
	return OutputJSON
}

// knownArrayKeys lists common wrapper keys that contain item arrays.
var knownArrayKeys = []string{"gifts", "chats", "contacts", "messages", "items"}

// printIDs extracts idKey from each item in the result and prints one per line to stdout.
func printIDs(result any, idKey string) {
	if idKey == "" {
		idKey = "id"
	}

	rMap, ok := result.(map[string]any)
	if !ok {
		// Not a map — try to print a single value
		_, _ = fmt.Fprintln(os.Stdout, fmt.Sprint(result))
		return
	}

	// Find the items array
	var items []any
	for _, key := range knownArrayKeys {
		if arr, ok := rMap[key].([]any); ok {
			items = arr
			break
		}
	}

	if items == nil {
		// No array found — try to extract idKey from the map itself
		if val, ok := rMap[idKey]; ok {
			_, _ = fmt.Fprintln(os.Stdout, formatIDValue(val))
		}
		return
	}

	for _, item := range items {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		val, ok := itemMap[idKey]
		if !ok {
			// Fallback to numeric "id"
			val, ok = itemMap["id"]
			if !ok {
				continue
			}
		}
		_, _ = fmt.Fprintln(os.Stdout, formatIDValue(val))
	}
}

// formatIDValue formats an ID value for output.
func formatIDValue(val any) string {
	switch v := val.(type) {
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	case string:
		return v
	default:
		return fmt.Sprint(v)
	}
}
