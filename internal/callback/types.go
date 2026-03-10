package callback

// Event is the payload sent to the configured callback URL.
type Event struct {
	EventType string `json:"eventType"` // "new_post", "edit_post"
	Post      Post   `json:"post"`
}

// Post represents a Telegram channel post in TGStat-like format.
type Post struct {
	ID            int            `json:"id"`
	Date          int            `json:"date"`
	Views         int            `json:"views,omitempty"`
	Link          string         `json:"link,omitempty"`
	ChannelID     string         `json:"channelId,omitempty"`
	ForwardedFrom string         `json:"forwardedFrom,omitempty"`
	IsDeleted     bool           `json:"isDeleted"`
	GroupID       *int64         `json:"groupId"`
	Text          string         `json:"text"`
	PostAuthor    string         `json:"postAuthor,omitempty"`
	Media         map[string]any `json:"media,omitempty"`
}
