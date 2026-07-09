// Package types provides common types for Telegram client send operations.
package types // revive:disable:var-naming

import "fmt"

// SendMessageParams holds parameters for SendMessage.
type SendMessageParams struct {
	PeerInfo
	Message string `json:"message" validate:"required"`
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
	if err := ValidateLatitude(p.Latitude); err != nil {
		return err
	}
	return ValidateLongitude(p.Longitude)
}

func (SendLocationParams) SchemaPropertyHints() map[string]map[string]any {
	return map[string]map[string]any{
		"latitude":  {"minimum": -90, "maximum": 90},
		"longitude": {"minimum": -180, "maximum": 180},
	}
}

// SendLocationResult is the result of SendLocation.
type SendLocationResult struct {
	ID        int64   `json:"id"`
	Date      int64   `json:"date"`
	Peer      string  `json:"peer"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// SendPhotoParams holds parameters for SendPhoto.
type SendPhotoParams struct {
	PeerInfo
	File    string `json:"file" validate:"required"`
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
	PeerInfo
	Phone     string `json:"phone" validate:"required"`
	FirstName string `json:"firstName" validate:"required"`
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
	PeerInfo
	File    string `json:"file" validate:"required"`
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
	Text string `json:"text" validate:"required"`
}

// SendPollParams holds parameters for SendPoll.
type SendPollParams struct {
	PeerInfo
	Question   string       `json:"question" validate:"required"`
	Options    []PollOption `json:"options"`
	Anonymous  bool         `json:"anonymous,omitempty"`
	Quiz       bool         `json:"quiz,omitempty"`
	CorrectIdx int          `json:"correctIdx,omitempty"`
}

// Validate validates SendPollParams.
func (p SendPollParams) Validate() error {
	if len(p.Options) < 2 {
		return fmt.Errorf("at least 2 options are required")
	}
	if len(p.Options) > 10 {
		return fmt.Errorf("maximum 10 options allowed")
	}
	return nil
}

func (SendPollParams) SchemaPropertyHints() map[string]map[string]any {
	return map[string]map[string]any{
		"options": {"minItems": 2, "maxItems": 10},
	}
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
	File    string `json:"file" validate:"required"`
	Caption string `json:"caption,omitempty"`
}

// SendVideoResult is the result of SendVideo.
type SendVideoResult struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"`
	Peer    string `json:"peer"`
	Caption string `json:"caption,omitempty"`
}

// SendVoiceParams holds parameters for SendVoice.
type SendVoiceParams struct {
	NoValidation
	Peer     string `json:"peer" validate:"required"`
	File     string `json:"file" validate:"required"` // Path to voice file (OGG/OPUS)
	Duration int    `json:"duration,omitempty"`       // Duration in seconds
	Caption  string `json:"caption,omitempty"`
}

// SendVoiceResult is the result of SendVoice.
type SendVoiceResult struct {
	ID       int64  `json:"id"`
	Date     int64  `json:"date"`
	Peer     string `json:"peer"`
	Duration int    `json:"duration,omitempty"`
}

// SendVideoNoteParams holds parameters for SendVideoNote.
type SendVideoNoteParams struct {
	NoValidation
	Peer     string `json:"peer" validate:"required"`
	File     string `json:"file" validate:"required"` // Path to video file
	Duration int    `json:"duration,omitempty"`       // Duration in seconds
	Length   int    `json:"length,omitempty"`         // Video width/height (square)
}

// SendVideoNoteResult is the result of SendVideoNote.
type SendVideoNoteResult struct {
	ID       int64  `json:"id"`
	Date     int64  `json:"date"`
	Peer     string `json:"peer"`
	Duration int    `json:"duration,omitempty"`
}

// SendGIFParams holds parameters for SendGIF.
type SendGIFParams struct {
	NoValidation
	Peer    string `json:"peer" validate:"required"`
	File    string `json:"file" validate:"required"` // Path to GIF file
	Caption string `json:"caption,omitempty"`
}

// SendGIFResult is the result of SendGIF.
type SendGIFResult struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"`
	Peer    string `json:"peer"`
	Caption string `json:"caption,omitempty"`
}

// ErrRequiresStickerOrFile is returned when neither stickerId nor file is provided.
var ErrRequiresStickerOrFile = fmt.Errorf("stickerId or file is required")

// SendStickerParams holds parameters for SendSticker.
type SendStickerParams struct {
	Peer      string `json:"peer" validate:"required"`
	StickerID string `json:"stickerId,omitempty"` // Sticker file_id or short_name
	File      string `json:"file,omitempty"`      // Path to sticker file (WEBP)
}

// Validate validates SendStickerParams.
func (p SendStickerParams) Validate() error {
	if p.StickerID == "" && p.File == "" {
		return ErrRequiresStickerOrFile
	}
	return nil
}

func (SendStickerParams) SchemaRules() map[string]any {
	return eitherRequiredSchema("stickerId", "file")
}

// SendStickerResult is the result of SendSticker.
type SendStickerResult struct {
	ID   int64  `json:"id"`
	Date int64  `json:"date"`
	Peer string `json:"peer"`
}

// SendDiceParams holds parameters for SendDice.
type SendDiceParams struct {
	PeerInfo
	Emoticon string `json:"emoticon,omitempty"`
	ReplyTo  int64  `json:"replyTo,omitempty"`
}

// SendDiceResult is the result of SendDice.
type SendDiceResult struct {
	ID       int64  `json:"id"`
	Date     int64  `json:"date"`
	Peer     string `json:"peer"`
	Value    int    `json:"value"`
	Emoticon string `json:"emoticon"`
}

// GetStickerPacksParams holds parameters for GetStickerPacks.
type GetStickerPacksParams struct {
	NoValidation
	// No required params - returns all sticker packs
}

// StickerPack represents a sticker pack.
type StickerPack struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	ShortName string `json:"shortName"`
	Count     int    `json:"count"`
	Animated  bool   `json:"animated,omitempty"`
	Video     bool   `json:"video,omitempty"`
}

// GetStickerPacksResult is the result of GetStickerPacks.
type GetStickerPacksResult struct {
	Packs []StickerPack `json:"packs"`
	Count int           `json:"count"`
}
