// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"agent-telegram/telegram/types"
)

// GetUpdatesHandler returns a handler for get_updates requests.
func GetUpdatesHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.GetUpdatesParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		// Set defaults
		if p.Limit <= 0 {
			p.Limit = 10
		}
		if p.Limit > 100 {
			p.Limit = 100
		}

		// Get updates - fetch more if filtering to get enough matching results
		fetchLimit := p.Limit
		if p.Peer != "" || p.Username != "" {
			fetchLimit = 100 // Get more to filter from
		}
		updates := client.GetUpdates(fetchLimit)

		// Filter by peer if specified
		if p.Peer != "" || p.Username != "" {
			updates = filterByPeer(updates, p.Peer, p.Username)
			// Limit to requested count after filtering
			if len(updates) > p.Limit {
				updates = updates[:p.Limit]
			}
		}

		return types.GetUpdatesResult{
			Updates: updates,
			Count:   len(updates),
		}, nil
	}
}

// filterByPeer filters updates by peer or username.
func filterByPeer(updates []types.StoredUpdate, peer, username string) []types.StoredUpdate {
	var filtered []types.StoredUpdate

	// Normalize filter - use peer if set, otherwise username
	filterPeer := peer
	if filterPeer == "" {
		filterPeer = username
	}
	if filterPeer == "" {
		return updates
	}

	// Strip @ prefix for comparison
	filterValue := strings.TrimPrefix(filterPeer, "@")

	for _, u := range updates {
		// Check if peer matches in data
		if peerMatches(u.Data, filterValue) {
			filtered = append(filtered, u)
		}
	}

	return filtered
}

// peerMatches checks if an update's peer matches the filter value.
func peerMatches(data map[string]any, filterValue string) bool {
	// Try to match against peer ID (format: user:123, chat:123, channel:123)
	if peer, ok := data["peer"].(string); ok {
		// If filter is numeric, compare with ID portion
		if isNumeric(filterValue) {
			parts := strings.Split(peer, ":")
			if len(parts) == 2 && parts[1] == filterValue {
				return true
			}
		}
	}

	// Try to match against peer_name (display name or username)
	if peerName, ok := data["peer_name"].(string); ok {
		// Case-insensitive partial match for username
		if strings.Contains(strings.ToLower(peerName), strings.ToLower(filterValue)) {
			return true
		}
	}

	// Check inside message data for peer info
	if msg, ok := data["message"].(map[string]any); ok {
		// Check peerId field in message
		if peerID, ok := msg["peerId"].(string); ok {
			// If filter is numeric, compare with ID portion
			if isNumeric(filterValue) {
				parts := strings.Split(peerID, ":")
				if len(parts) == 2 && parts[1] == filterValue {
					return true
				}
			}
		}
	}

	return false
}

// isNumeric checks if a string is a number.
func isNumeric(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

