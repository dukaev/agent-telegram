package cliutil

import (
	"fmt"
	"strconv"
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

// FilterExpression represents a single filter condition (e.g., "stars>1000").
type FilterExpression struct {
	Key      string
	Operator string
	Value    string
}

// FilterExpressions is a list of filter expressions that AND together.
type FilterExpressions []*FilterExpression

// operators ordered longest-first to avoid prefix matching issues.
var filterOperators = []string{">=", "<=", "!=", ">", "<", "="}

// ParseFilterExpressions parses filter strings like "stars>1000" or "type=channel".
func ParseFilterExpressions(filters []string) (FilterExpressions, error) {
	if len(filters) == 0 {
		return nil, nil
	}
	exprs := make(FilterExpressions, 0, len(filters))
	for _, f := range filters {
		expr, err := parseOneFilter(f)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}
	return exprs, nil
}

func parseOneFilter(s string) (*FilterExpression, error) {
	for _, op := range filterOperators {
		idx := strings.Index(s, op)
		if idx > 0 {
			return &FilterExpression{
				Key:      strings.TrimSpace(s[:idx]),
				Operator: op,
				Value:    strings.TrimSpace(s[idx+len(op):]),
			}, nil
		}
	}
	return nil, fmt.Errorf("invalid filter expression %q: must contain an operator (=, !=, >, <, >=, <=)", s)
}

// Apply filters the result, keeping only items that match all expressions.
// Works on wrapper maps with known array keys.
func (fe FilterExpressions) Apply(result any) any {
	if len(fe) == 0 {
		return result
	}

	rMap, ok := result.(map[string]any)
	if !ok {
		return result
	}

	// Find the items array in the wrapper
	for _, key := range knownArrayKeys {
		arr, ok := rMap[key].([]any)
		if !ok {
			continue
		}

		filtered := make([]map[string]any, 0, len(arr))
		for _, item := range arr {
			itemMap, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if fe.matchesAll(itemMap) {
				filtered = append(filtered, itemMap)
			}
		}

		return MapResponse(rMap, key, filtered)
	}

	// No array found — try filtering the map itself
	if fe.matchesAll(rMap) {
		return result
	}
	return nil
}

// matchesAll checks if an item matches all filter expressions.
func (fe FilterExpressions) matchesAll(item map[string]any) bool {
	for _, expr := range fe {
		if !expr.matches(item) {
			return false
		}
	}
	return true
}

// matches checks if a single item matches this filter expression.
func (expr *FilterExpression) matches(item map[string]any) bool {
	val, ok := item[expr.Key]
	if !ok {
		return false
	}

	// Try numeric comparison first
	if itemNum, ok := toFloat64(val); ok {
		if exprNum, err := strconv.ParseFloat(expr.Value, 64); err == nil {
			return compareNumbers(itemNum, expr.Operator, exprNum)
		}
	}

	// Fallback to case-insensitive string comparison
	itemStr := strings.ToLower(fmt.Sprint(val))
	exprStr := strings.ToLower(expr.Value)
	return compareStrings(itemStr, expr.Operator, exprStr)
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int64:
		return float64(n), true
	case int:
		return float64(n), true
	}
	return 0, false
}

func compareNumbers(a float64, op string, b float64) bool {
	switch op {
	case "=":
		return a == b
	case "!=":
		return a != b
	case ">":
		return a > b
	case "<":
		return a < b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	}
	return false
}

func compareStrings(a, op, b string) bool {
	switch op {
	case "=":
		return a == b
	case "!=":
		return a != b
	case ">":
		return a > b
	case "<":
		return a < b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	}
	return false
}
