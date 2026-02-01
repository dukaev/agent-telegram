package cliutil

import (
	"strings"
)

// FilterOptions defines filtering criteria for list results.
type FilterOptions struct {
	Search string // Search query (case-insensitive)
	Type   string // Filter by type (e.g., "user", "chat", "channel", "bot")
}

// FieldExtractor extracts a field value from an item.
type FieldExtractor func(item map[string]any) string

// TypeFilter determines if an item matches the specified type.
type TypeFilter func(item map[string]any, itemType string) bool

// FilterResult defines the result of filtering.
type FilterResult struct {
	Items  []map[string]any `json:"items"`
	Count  int              `json:"count"`
	Total  int              `json:"total,omitempty"`
	Limit  int              `json:"limit,omitempty"`
	Offset int              `json:"offset,omitempty"`
}

// FilterItems filters a list of items based on the provided options.
func FilterItems(items []map[string]any, opts FilterOptions) FilterResult {
	result := FilterResult{
		Items: make([]map[string]any, 0),
	}

	searchLower := strings.ToLower(opts.Search)

	for _, item := range items {
		if !matchesFilters(item, opts, searchLower) {
			continue
		}
		result.Items = append(result.Items, item)
	}

	result.Count = len(result.Items)
	result.Total = len(items)

	return result
}

// matchesFilters checks if an item matches all filter criteria.
func matchesFilters(item map[string]any, opts FilterOptions, searchLower string) bool {
	// Type filter
	if opts.Type != "" && !matchesTypeFilter(item, opts.Type) {
		return false
	}

	// Search filter
	if opts.Search != "" && !matchesSearchFilter(item, searchLower) {
		return false
	}

	return true
}

// matchesTypeFilter checks if an item matches the type filter.
func matchesTypeFilter(item map[string]any, filterType string) bool {
	itemType := ExtractString(item, "type")
	if itemType == "" {
		return false
	}

	// Special case for "bot" type - filter users that are bots
	if filterType == "bot" {
		if itemType != "user" {
			return false
		}
		isBot, _ := item["bot"].(bool)
		return isBot
	}

	return itemType == filterType
}

// matchesSearchFilter checks if an item matches the search filter.
func matchesSearchFilter(item map[string]any, searchLower string) bool {
	// Check common searchable fields
	searchableFields := []string{"title", "username", "peer", "name", "first_name", "last_name"}

	for _, field := range searchableFields {
		value := ExtractString(item, field)
		if value != "" && strings.Contains(strings.ToLower(value), searchLower) {
			return true
		}
	}

	return false
}

// MapResponse transforms a response map with a items array into a FilterResult format.
func MapResponse(response map[string]any, itemsKey string, filteredItems []map[string]any) map[string]any {
	result := make(map[string]any)
	result[itemsKey] = filteredItems

	// Copy common fields
	if limit, ok := response["limit"]; ok {
		result["limit"] = limit
	}
	if offset, ok := response["offset"]; ok {
		result["offset"] = offset
	}
	result["count"] = len(filteredItems)
	if total, ok := response["total"]; ok {
		result["total"] = total
	} else if count, ok := response["count"]; ok {
		result["total"] = count
	}

	return result
}

// ContainsAny checks if any of the strings contain the search term (case-insensitive).
func ContainsAny(search string, targets ...string) bool {
	searchLower := strings.ToLower(search)
	for _, target := range targets {
		if target != "" && strings.Contains(strings.ToLower(target), searchLower) {
			return true
		}
	}
	return false
}
