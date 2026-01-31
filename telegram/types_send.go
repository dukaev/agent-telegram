// Package telegram provides common types for Telegram client send operations.
package telegram

// SendMessageParams holds parameters for SendMessage.
type SendMessageParams struct {
	Peer    string `json:"peer"`
	Message string `json:"message"`
}

// SendMessageResult is the result of SendMessage.
type SendMessageResult struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"`
	Message string `json:"message"`
	Peer    string `json:"peer"`
}

// SendLocationParams holds parameters for SendLocation.
type SendLocationParams struct {
	Peer     string  `json:"peer"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// SendLocationResult is the result of SendLocation.
type SendLocationResult struct {
	ID       int64   `json:"id"`
	Date     int64   `json:"date"`
	Peer     string  `json:"peer"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// SendPhotoParams holds parameters for SendPhoto.
type SendPhotoParams struct {
	Peer    string `json:"peer"`
	File    string `json:"file"`
	Caption string `json:"caption,omitempty"`
}

// SendPhotoResult is the result of SendPhoto.
type SendPhotoResult struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"`
	Peer    string `json:"peer"`
	Caption string `json:"caption,omitempty"`
}

// SendContactParams holds parameters for SendContact.
type SendContactParams struct {
	Peer      string `json:"peer"`
	Phone     string `json:"phone"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName,omitempty"`
}

// SendContactResult is the result of SendContact.
type SendContactResult struct {
	ID    int64  `json:"id"`
	Date  int64  `json:"date"`
	Peer  string `json:"peer"`
	Phone string `json:"phone"`
}

// SendFileParams holds parameters for SendFile.
type SendFileParams struct {
	Peer    string `json:"peer"`
	File    string `json:"file"`
	Caption string `json:"caption,omitempty"`
}

// SendFileResult is the result of SendFile.
type SendFileResult struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"`
	Peer    string `json:"peer"`
	Caption string `json:"caption,omitempty"`
}

// PollOption represents a poll option.
type PollOption struct {
	Text string `json:"text"`
}

// SendPollParams holds parameters for SendPoll.
type SendPollParams struct {
	Peer       string       `json:"peer"`
	Question   string       `json:"question"`
	Options    []PollOption `json:"options"`
	Anonymous  bool         `json:"anonymous,omitempty"`
	Quiz       bool         `json:"quiz,omitempty"`
	CorrectIdx int          `json:"correctIdx,omitempty"`
}

// SendPollResult is the result of SendPoll.
type SendPollResult struct {
	ID       int64  `json:"id"`
	Date     int64  `json:"date"`
	Peer     string `json:"peer"`
	Question string `json:"question"`
}

// SendVideoParams holds parameters for SendVideo.
type SendVideoParams struct {
	Peer    string `json:"peer"`
	File    string `json:"file"`
	Caption string `json:"caption,omitempty"`
}

// SendVideoResult is the result of SendVideo.
type SendVideoResult struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"`
	Peer    string `json:"peer"`
	Caption string `json:"caption,omitempty"`
}
