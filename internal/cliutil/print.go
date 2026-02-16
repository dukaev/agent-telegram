// Package cliutil provides generic filtering utilities for CLI commands.
package cliutil

import (
	"fmt"
	"os"
	"time"
)

// PrintSuccessSummary prints a success summary from a result map.
func PrintSuccessSummary(result any, message string) {
	r, ok := result.(map[string]any)
	if !ok {
		return
	}
	if success, ok := r["success"].(bool); ok && success {
		fmt.Fprintf(os.Stderr, "%s\n", message)
	}
}

// PrintSuccessWithDuration prints a success summary with operation duration.
func PrintSuccessWithDuration(result any, message string, d time.Duration) {
	r, ok := result.(map[string]any)
	if !ok {
		return
	}
	if success, ok := r["success"].(bool); ok && success {
		fmt.Fprintf(os.Stderr, "%s (%s)\n", message, formatDuration(d))
	}
}

// formatDuration formats a duration for display (e.g. "342ms", "1.2s").
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

// PrintResultField prints a field from result map if it exists.
func PrintResultField(result any, key string, format string, args ...any) {
	r, ok := result.(map[string]any)
	if !ok {
		return
	}
	switch v := r[key].(type) {
	case string:
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, format, v)
		} else {
			fmt.Fprintf(os.Stderr, format, append([]any{v}, args...)...)
		}
	case float64:
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, format, int64(v))
		} else {
			newArgs := append([]any{int64(v)}, args...)
			fmt.Fprintf(os.Stderr, format, newArgs...)
		}
	case int64:
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, format, v)
		} else {
			fmt.Fprintf(os.Stderr, format, append([]any{v}, args...)...)
		}
	}
}

// PrintInviteLinkSummary prints invite link summary.
func PrintInviteLinkSummary(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		return
	}
	PrintResultField(r, "link", "Invite link: %s\n")
	PrintResultField(r, "usage", "Usage: %d\n")
	PrintResultField(r, "usageLimit", "Usage limit: %d\n")
}

// PrintResultCount prints count from result map.
func PrintResultCount(result any, key string, label string) {
	r, ok := result.(map[string]any)
	if !ok {
		return
	}
	if count, ok := r[key].(float64); ok {
		fmt.Fprintf(os.Stderr, "%s: %d\n", label, int(count))
	}
}

// IterateItems iterates over items in result map and calls fn for each.
func IterateItems(result any, key string, fn func(map[string]any)) {
	r, ok := result.(map[string]any)
	if !ok {
		return
	}
	items, ok := r[key].([]any)
	if !ok {
		return
	}
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			fn(m)
		}
	}
}

// ExtractStringValue safely extracts a string value from map.
func ExtractStringValue(item map[string]any, key string) string {
	if s, ok := item[key].(string); ok {
		return s
	}
	return ""
}

// FormatName formats a name from first and last name fields.
func FormatName(item map[string]any) string {
	firstName := ExtractStringValue(item, "firstName")
	lastName := ExtractStringValue(item, "lastName")
	if lastName != "" {
		return firstName + " " + lastName
	}
	return firstName
}

// PrintParticipants prints participant list from result.
func PrintParticipants(result any, firstNameDefault string, peerDefault string) {
	PrintResultCount(result, "count", "Found %d participant(s)")
	IterateItems(result, "participants", func(participant map[string]any) {
		name := FormatName(participant)
		if name == "" {
			name = firstNameDefault
		}
		peer := ExtractStringValue(participant, "peer")
		if peer == "" {
			peer = peerDefault
		}
		fmt.Fprintf(os.Stderr, "  - %s (%s)\n", name, peer)
	})
}

// PrintBanned prints banned users list from result.
func PrintBanned(result any, firstNameDefault string, peerDefault string) {
	PrintResultCount(result, "count", "Found %d banned user(s)")
	IterateItems(result, "banned", func(user map[string]any) {
		name := FormatName(user)
		if name == "" {
			name = firstNameDefault
		}
		peer := ExtractStringValue(user, "peer")
		if peer == "" {
			peer = peerDefault
		}
		fmt.Fprintf(os.Stderr, "  - %s (%s)\n", name, peer)
	})
}

// PrintAdmins prints admins list from result.
func PrintAdmins(result any, firstNameDefault string, peerDefault string) {
	PrintResultCount(result, "count", "Found %d admin(s)")
	IterateItems(result, "admins", func(admin map[string]any) {
		name := FormatName(admin)
		if name == "" {
			name = firstNameDefault
		}
		peer := ExtractStringValue(admin, "peer")
		if peer == "" {
			peer = peerDefault
		}
		suffix := ""
		if creator, ok := admin["creator"].(bool); ok && creator {
			suffix = " (Creator)"
		}
		fmt.Fprintf(os.Stderr, "  - %s (%s)%s\n", name, peer, suffix)
	})
}

// PrintTopics prints topics list from result.
func PrintTopics(result any, defaultValue string) {
	PrintResultCount(result, "count", "Found %d topic(s)")
	IterateItems(result, "topics", func(topic map[string]any) {
		title := ExtractStringValue(topic, "title")
		if title == "" {
			title = defaultValue
		}
		id := defaultValue
		if i, ok := topic["id"].(float64); ok {
			id = fmt.Sprintf("%.0f", i)
		}
		fmt.Fprintf(os.Stderr, "  - [%s] %s\n", id, title)
	})
}
