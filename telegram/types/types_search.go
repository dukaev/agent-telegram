// Package types provides common types for Telegram client search operations.
package types // revive:disable:var-naming

// SearchResult represents a single search result.
type SearchResult struct {
	ID       int64         `json:"id"`
	Date     int64         `json:"date"`
	Text     string        `json:"text,omitempty"`
	Peer     string        `json:"peer,omitempty"`
	FromID   string        `json:"fromId,omitempty"`
	FromName string        `json:"fromName,omitempty"`
	Media    map[string]any `json:"media,omitempty"`
}

// SearchGlobalParams holds parameters for SearchGlobal.
type SearchGlobalParams struct {
	Query string `json:"query"`
	Type  string `json:"type,omitempty"` // bots, users, chats, channels, or empty for all
	Limit int    `json:"limit,omitempty"`
}

// SearchGlobalResult is the result of SearchGlobal.
type SearchGlobalResult struct {
	Query   string        `json:"query"`
	Type    string        `json:"type,omitempty"`
	Results []SearchResult `json:"results"`
	Count   int           `json:"count"`
}

// SearchInChatParams holds parameters for SearchInChat.
type SearchInChatParams struct {
	Peer   string `json:"peer"`
	Query  string `json:"query"`
	Type   string `json:"type,omitempty"` // text, photos, videos, documents, links, audio, voice
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// SearchInChatResult is the result of SearchInChat.
type SearchInChatResult struct {
	Peer     string          `json:"peer"`
	Query    string          `json:"query"`
	Type     string          `json:"type,omitempty"`
	Messages []MessageResult `json:"messages"`
	Count    int             `json:"count"`
	Total    int             `json:"total"`
	Limit    int             `json:"limit"`
	Offset   int             `json:"offset"`
}
