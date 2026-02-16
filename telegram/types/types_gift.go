// Package types provides common types for Telegram client gift operations.
package types // revive:disable:var-naming

import "fmt"

// GetStarGiftsParams holds parameters for GetStarGifts.
type GetStarGiftsParams struct {
	Limit int `json:"limit,omitempty"`
}

// Validate validates GetStarGiftsParams.
func (p GetStarGiftsParams) Validate() error {
	return nil
}

// GiftItem represents a star gift from the catalog.
type GiftItem struct {
	ID                  int64  `json:"id"`
	Stars               int64  `json:"stars"`
	ConvertStars        int64  `json:"convertStars"`
	Limited             bool   `json:"limited,omitempty"`
	SoldOut             bool   `json:"soldOut,omitempty"`
	Birthday            bool   `json:"birthday,omitempty"`
	AvailabilityRemains int    `json:"availabilityRemains,omitempty"`
	AvailabilityTotal   int    `json:"availabilityTotal,omitempty"`
	Title               string `json:"title,omitempty"`
	UpgradeStars        int64  `json:"upgradeStars,omitempty"`
	Slug                string `json:"slug,omitempty"`
}

// GetStarGiftsResult is the result of GetStarGifts.
type GetStarGiftsResult struct {
	Gifts []GiftItem `json:"gifts"`
	Count int        `json:"count"`
}

// SendStarGiftParams holds parameters for SendStarGift.
type SendStarGiftParams struct {
	Peer     string `json:"peer" validate:"required"`
	Slug     string `json:"slug" validate:"required"`
	Price    int64  `json:"price" validate:"required"`
	Duration int    `json:"duration,omitempty"`
}

// Validate validates SendStarGiftParams.
func (p SendStarGiftParams) Validate() error {
	return ValidateStruct(p)
}

// SendStarGiftResult is the result of SendStarGift.
type SendStarGiftResult struct {
	Success bool `json:"success"`
}

// GetSavedGiftsParams holds parameters for GetSavedGifts.
type GetSavedGiftsParams struct {
	Peer   string `json:"peer,omitempty"`
	Offset string `json:"offset,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// Validate validates GetSavedGiftsParams.
func (p GetSavedGiftsParams) Validate() error {
	return nil
}

// SavedGiftItem represents a saved star gift.
type SavedGiftItem struct {
	MsgID         int    `json:"msgId,omitempty"`
	SavedID       int64  `json:"savedId,omitempty"`
	GiftID        int64  `json:"giftId"`
	FromID        string `json:"fromId,omitempty"`
	Date          int    `json:"date"`
	Stars         int64  `json:"stars"`
	ConvertStars  int64  `json:"convertStars,omitempty"`
	TransferStars int64  `json:"transferStars,omitempty"`
	Message       string `json:"message,omitempty"`
	NameHidden    bool   `json:"nameHidden,omitempty"`
	Unsaved       bool   `json:"unsaved,omitempty"`
	CanUpgrade    bool   `json:"canUpgrade,omitempty"`
	Slug          string `json:"slug,omitempty"`
	Title         string `json:"title,omitempty"`
}

// GetSavedGiftsResult is the result of GetSavedGifts.
type GetSavedGiftsResult struct {
	Gifts      []SavedGiftItem `json:"gifts"`
	Count      int             `json:"count"`
	NextOffset string          `json:"nextOffset,omitempty"`
}

// TransferStarGiftParams holds parameters for TransferStarGift.
type TransferStarGiftParams struct {
	Peer  string `json:"peer" validate:"required"`
	MsgID int    `json:"msgId,omitempty"`
	Slug  string `json:"slug,omitempty"`
}

// Validate validates TransferStarGiftParams.
func (p TransferStarGiftParams) Validate() error {
	if p.Peer == "" {
		return fmt.Errorf("peer is required")
	}
	if p.MsgID == 0 && p.Slug == "" {
		return fmt.Errorf("either msgId or slug is required")
	}
	return nil
}

// TransferStarGiftResult is the result of TransferStarGift.
type TransferStarGiftResult struct {
	Success bool `json:"success"`
}

// ConvertStarGiftParams holds parameters for ConvertStarGift.
type ConvertStarGiftParams struct {
	MsgID int    `json:"msgId,omitempty"`
	Slug  string `json:"slug,omitempty"`
}

// Validate validates ConvertStarGiftParams.
func (p ConvertStarGiftParams) Validate() error {
	if p.MsgID == 0 && p.Slug == "" {
		return fmt.Errorf("either msgId or slug is required")
	}
	return nil
}

// ConvertStarGiftResult is the result of ConvertStarGift.
type ConvertStarGiftResult struct {
	Success bool `json:"success"`
}
