// Package telegram provides common types for Telegram client send operations.
package telegram

import "fmt"

// SendMessageParams holds parameters for SendMessage.
type SendMessageParams struct {
	PeerInfo
	Message string `json:"message"`
}

// Validate validates SendMessageParams.
func (p SendMessageParams) Validate() error {
	if err := p.ValidatePeer(); err != nil {
		return err
	}
	if p.Message == "" {
		return fmt.Errorf("message is required")
	}
	return nil
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
	PeerInfo
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Validate validates SendLocationParams.
func (p SendLocationParams) Validate() error {
	if err := p.ValidatePeer(); err != nil {
		return err
	}
	if p.Latitude < -90 || p.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if p.Longitude < -180 || p.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	return nil
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
	PeerInfo
	File    string `json:"file"`
	Caption string `json:"caption,omitempty"`
}

// Validate validates SendPhotoParams.
func (p SendPhotoParams) Validate() error {
	if err := p.ValidatePeer(); err != nil {
		return err
	}
	if p.File == "" {
		return fmt.Errorf("file is required")
	}
	return nil
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
	PeerInfo
	Phone    string `json:"phone"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName,omitempty"`
}

// Validate validates SendContactParams.
func (p SendContactParams) Validate() error {
	if err := p.ValidatePeer(); err != nil {
		return err
	}
	if p.Phone == "" {
		return fmt.Errorf("phone is required")
	}
	if p.FirstName == "" {
		return fmt.Errorf("firstName is required")
	}
	return nil
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
	PeerInfo
	File    string `json:"file"`
	Caption string `json:"caption,omitempty"`
}

// Validate validates SendFileParams.
func (p SendFileParams) Validate() error {
	if err := p.ValidatePeer(); err != nil {
		return err
	}
	if p.File == "" {
		return fmt.Errorf("file is required")
	}
	return nil
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
	PeerInfo
	Question    string       `json:"question"`
	Options     []PollOption `json:"options"`
	Anonymous   bool         `json:"anonymous,omitempty"`
	Quiz        bool         `json:"quiz,omitempty"`
	CorrectIdx  int          `json:"correctIdx,omitempty"`
}

// Validate validates SendPollParams.
func (p SendPollParams) Validate() error {
	if err := p.ValidatePeer(); err != nil {
		return err
	}
	if p.Question == "" {
		return fmt.Errorf("question is required")
	}
	if len(p.Options) < 2 {
		return fmt.Errorf("at least 2 options are required")
	}
	if len(p.Options) > 10 {
		return fmt.Errorf("maximum 10 options allowed")
	}
	return nil
}

// ValidateForQuiz validates SendPollParams for quiz mode.
func (p SendPollParams) ValidateForQuiz() error {
	if err := p.Validate(); err != nil {
		return err
	}
	if p.CorrectIdx < 0 || p.CorrectIdx >= len(p.Options) {
		return fmt.Errorf("correctIdx must be between 0 and %d", len(p.Options)-1)
	}
	return nil
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
	PeerInfo
	File    string `json:"file"`
	Caption string `json:"caption,omitempty"`
}

// Validate validates SendVideoParams.
func (p SendVideoParams) Validate() error {
	if err := p.ValidatePeer(); err != nil {
		return err
	}
	if p.File == "" {
		return fmt.Errorf("file is required")
	}
	return nil
}

// SendVideoResult is the result of SendVideo.
type SendVideoResult struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"`
	Peer    string `json:"peer"`
	Caption string `json:"caption,omitempty"`
}
