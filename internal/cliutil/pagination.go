// Package cliutil provides CLI utility functions.
package cliutil

// Pagination limit constants.
const (
	// DefaultLimitSmall is for chats, messages, updates.
	DefaultLimitSmall = 10
	// DefaultLimitMedium is for search results.
	DefaultLimitMedium = 20
	// DefaultLimitLarge is for contacts.
	DefaultLimitLarge = 50
	// DefaultLimitMax is for topics, participants.
	DefaultLimitMax = 100

	// MaxLimitStandard is the maximum for most commands.
	MaxLimitStandard = 100
	// MaxLimitParticipants is the maximum for admins, banned, participants.
	MaxLimitParticipants = 200
)

// PaginationConfig defines pagination parameters for validation.
type PaginationConfig struct {
	MaxLimit int
}

// Pagination holds validated pagination parameters.
type Pagination struct {
	Limit  int
	Offset int
}

// NewPagination creates a Pagination with validated values.
func NewPagination(limit, offset int, cfg PaginationConfig) Pagination {
	// Validate limit
	if limit < 1 {
		limit = 1
	}
	if cfg.MaxLimit > 0 && limit > cfg.MaxLimit {
		limit = cfg.MaxLimit
	}

	// Validate offset
	if offset < 0 {
		offset = 0
	}

	return Pagination{
		Limit:  limit,
		Offset: offset,
	}
}

// ToParams adds pagination to a params map.
func (p Pagination) ToParams(params map[string]any, includeOffset bool) {
	params["limit"] = p.Limit
	if includeOffset {
		params["offset"] = p.Offset
	}
}
