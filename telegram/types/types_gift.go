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
	GiftID   int64  `json:"giftId,omitempty"`
	Name     string `json:"name,omitempty"`
	Message  string `json:"message,omitempty"`
	HideName bool   `json:"hideName,omitempty"`
}

// Validate validates SendStarGiftParams.
func (p SendStarGiftParams) Validate() error {
	if p.Peer == "" {
		return fmt.Errorf("peer is required")
	}
	if p.GiftID == 0 && p.Name == "" {
		return fmt.Errorf("either giftId or name is required")
	}
	return nil
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
	ResellStars   int64  `json:"resellStars,omitempty"`
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

// UpdateGiftPriceParams holds parameters for UpdateGiftPrice.
type UpdateGiftPriceParams struct {
	MsgID int    `json:"msgId,omitempty"`
	Slug  string `json:"slug,omitempty"`
	Price int64  `json:"price" validate:"required"`
}

// Validate validates UpdateGiftPriceParams.
func (p UpdateGiftPriceParams) Validate() error {
	if p.MsgID == 0 && p.Slug == "" {
		return fmt.Errorf("either msgId or slug is required")
	}
	return ValidateStruct(p)
}

// UpdateGiftPriceResult is the result of UpdateGiftPrice.
type UpdateGiftPriceResult struct {
	Success bool `json:"success"`
}

// GetBalanceParams holds parameters for GetBalance.
type GetBalanceParams struct{}

// Validate validates GetBalanceParams.
func (p GetBalanceParams) Validate() error {
	return nil
}

// GetBalanceResult is the result of GetBalance.
type GetBalanceResult struct {
	Stars int64 `json:"stars"`
	Nanos int   `json:"nanos,omitempty"`
	Ton   int64 `json:"ton"`
}

// OfferGiftParams holds parameters for OfferGift.
type OfferGiftParams struct {
	Peer     string `json:"peer" validate:"required"`
	Slug     string `json:"slug" validate:"required"`
	Price    int64  `json:"price" validate:"required"`
	Duration int    `json:"duration,omitempty"`
}

// Validate validates OfferGiftParams.
func (p OfferGiftParams) Validate() error {
	return ValidateStruct(p)
}

// OfferGiftResult is the result of OfferGift.
type OfferGiftResult struct {
	Success bool `json:"success"`
}

// GetResaleGiftsParams holds parameters for GetResaleGifts.
type GetResaleGiftsParams struct {
	GiftID      int64  `json:"giftId,omitempty"`
	Name        string `json:"name,omitempty"`
	Offset      string `json:"offset,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	SortByPrice bool   `json:"sortByPrice,omitempty"`
	SortByNum   bool   `json:"sortByNum,omitempty"`
	Model       string `json:"model,omitempty"`
	Pattern     string `json:"pattern,omitempty"`
	Backdrop    string `json:"backdrop,omitempty"`
}

// Validate validates GetResaleGiftsParams.
func (p GetResaleGiftsParams) Validate() error {
	if p.GiftID == 0 && p.Name == "" {
		return fmt.Errorf("either giftId or name is required")
	}
	return nil
}

// ResaleGiftItem represents a gift listed for resale.
type ResaleGiftItem struct {
	ID                 int64           `json:"id"`
	GiftID             int64           `json:"giftId"`
	Title              string          `json:"title"`
	Slug               string          `json:"slug"`
	Num                int             `json:"num"`
	OwnerName          string          `json:"ownerName,omitempty"`
	ResellStars        int64           `json:"resellStars,omitempty"`
	AvailabilityIssued int             `json:"availabilityIssued"`
	AvailabilityTotal  int             `json:"availabilityTotal"`
	Attributes         []GiftAttribute `json:"attributes,omitempty"`
}

// GetResaleGiftsResult is the result of GetResaleGifts.
type GetResaleGiftsResult struct {
	Gifts      []ResaleGiftItem `json:"gifts"`
	Count      int              `json:"count"`
	NextOffset string           `json:"nextOffset,omitempty"`
}

// BuyResaleGiftParams holds parameters for BuyResaleGift.
type BuyResaleGiftParams struct {
	Slug string `json:"slug" validate:"required"`
	Peer string `json:"peer,omitempty"`
}

// Validate validates BuyResaleGiftParams.
func (p BuyResaleGiftParams) Validate() error {
	return ValidateStruct(p)
}

// BuyResaleGiftResult is the result of BuyResaleGift.
type BuyResaleGiftResult struct {
	Success bool `json:"success"`
}

// GetGiftValueParams holds parameters for GetGiftValue.
type GetGiftValueParams struct {
	Slug string `json:"slug" validate:"required"`
}

// Validate validates GetGiftValueParams.
func (p GetGiftValueParams) Validate() error {
	return ValidateStruct(p)
}

// GetGiftValueResult is the result of GetGiftValue.
type GetGiftValueResult struct {
	Currency            string `json:"currency"`
	Value               int64  `json:"value"`
	InitialSaleDate     int    `json:"initialSaleDate"`
	InitialSaleStars    int64  `json:"initialSaleStars"`
	InitialSalePrice    int64  `json:"initialSalePrice"`
	LastSaleDate        int    `json:"lastSaleDate,omitempty"`
	LastSalePrice       int64  `json:"lastSalePrice,omitempty"`
	LastSaleOnFragment  bool   `json:"lastSaleOnFragment,omitempty"`
	ValueIsAverage      bool   `json:"valueIsAverage,omitempty"`
	FloorPrice          int64  `json:"floorPrice,omitempty"`
	AveragePrice        int64  `json:"averagePrice,omitempty"`
	ListedCount         int    `json:"listedCount,omitempty"`
	FragmentListedCount int    `json:"fragmentListedCount,omitempty"`
	FragmentListedURL   string `json:"fragmentListedUrl,omitempty"`
}

// GetGiftInfoParams holds parameters for GetGiftInfo.
type GetGiftInfoParams struct {
	Slug string `json:"slug" validate:"required"`
}

// Validate validates GetGiftInfoParams.
func (p GetGiftInfoParams) Validate() error {
	return ValidateStruct(p)
}

// GiftAttribute represents a gift attribute (model, pattern, backdrop).
type GiftAttribute struct {
	Type           string `json:"type"`
	Name           string `json:"name"`
	RarityPermille int    `json:"rarityPermille,omitempty"`
}

// GetGiftAttrsParams holds parameters for GetGiftAttrs.
type GetGiftAttrsParams struct {
	GiftID int64  `json:"giftId,omitempty"`
	Name   string `json:"name,omitempty"`
}

// Validate validates GetGiftAttrsParams.
func (p GetGiftAttrsParams) Validate() error {
	if p.GiftID == 0 && p.Name == "" {
		return fmt.Errorf("either giftId or name is required")
	}
	return nil
}

// GetGiftAttrsResult is the result of GetGiftAttrs.
type GetGiftAttrsResult struct {
	GiftID    int64           `json:"giftId"`
	Models    []GiftAttribute `json:"models"`
	Patterns  []GiftAttribute `json:"patterns"`
	Backdrops []GiftAttribute `json:"backdrops"`
	Count     int             `json:"count"`
}

// AcceptGiftOfferParams holds parameters for AcceptGiftOffer.
type AcceptGiftOfferParams struct {
	OfferMsgID int `json:"offerMsgId" validate:"required"`
}

// Validate validates AcceptGiftOfferParams.
func (p AcceptGiftOfferParams) Validate() error {
	return ValidateStruct(p)
}

// AcceptGiftOfferResult is the result of AcceptGiftOffer.
type AcceptGiftOfferResult struct {
	Success bool `json:"success"`
}

// DeclineGiftOfferParams holds parameters for DeclineGiftOffer.
type DeclineGiftOfferParams struct {
	OfferMsgID int `json:"offerMsgId" validate:"required"`
}

// Validate validates DeclineGiftOfferParams.
func (p DeclineGiftOfferParams) Validate() error {
	return ValidateStruct(p)
}

// DeclineGiftOfferResult is the result of DeclineGiftOffer.
type DeclineGiftOfferResult struct {
	Success bool `json:"success"`
}

// GetGiftInfoResult is the result of GetGiftInfo.
type GetGiftInfoResult struct {
	ID                 int64           `json:"id"`
	GiftID             int64           `json:"giftId"`
	Title              string          `json:"title"`
	Slug               string          `json:"slug"`
	Num                int             `json:"num"`
	OwnerID            string          `json:"ownerId,omitempty"`
	OwnerName          string          `json:"ownerName,omitempty"`
	OwnerAddress       string          `json:"ownerAddress,omitempty"`
	AvailabilityIssued int             `json:"availabilityIssued"`
	AvailabilityTotal  int             `json:"availabilityTotal"`
	ResellStars        int64           `json:"resellStars,omitempty"`
	GiftAddress        string          `json:"giftAddress,omitempty"`
	Attributes         []GiftAttribute `json:"attributes,omitempty"`
}
