// Package gift provides Telegram star gift operations.
package gift

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/gotd/td/tg"

	"agent-telegram/telegram/client"
	"agent-telegram/telegram/types"
)

// Client provides gift operations.
type Client struct {
	*client.BaseClient
}

// NewClient creates a new gift client.
func NewClient(tc client.ParentClient) *Client {
	return &Client{
		BaseClient: &client.BaseClient{Parent: tc},
	}
}

// giftNames provides fallback names for gifts without titles.
var giftNames = map[int64]string{
	5170145012310081615: "Heart",
	5170233102089322756: "Bear",
}

// resolveGiftByName resolves a gift name to its ID.
func (c *Client) resolveGiftByName(ctx context.Context, name string) (int64, error) {
	// Reverse lookup in giftNames map
	for id, n := range giftNames {
		if strings.EqualFold(n, name) {
			return id, nil
		}
	}

	// Fallback: fetch catalog and match by title
	result, err := c.API.PaymentsGetStarGifts(ctx, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve gift name: %w", err)
	}

	gifts, ok := result.AsModified()
	if !ok {
		return 0, fmt.Errorf("gift not found: %s", name)
	}

	for _, g := range gifts.Gifts {
		switch gift := g.(type) {
		case *tg.StarGift:
			title := gift.Title
			if title == "" {
				title = giftNames[gift.ID]
			}
			if strings.EqualFold(title, name) {
				return gift.ID, nil
			}
		case *tg.StarGiftUnique:
			if strings.EqualFold(gift.Title, name) {
				return gift.ID, nil
			}
		}
	}

	return 0, fmt.Errorf("gift not found: %s", name)
}

// resolveGiftInput creates an InputSavedStarGiftClass from msgID or slug.
func resolveGiftInput(msgID int, slug string) tg.InputSavedStarGiftClass {
	if slug != "" {
		return &tg.InputSavedStarGiftSlug{Slug: slug}
	}
	return &tg.InputSavedStarGiftUser{MsgID: msgID}
}

// GetStarGifts returns the catalog of available star gifts.
func (c *Client) GetStarGifts(ctx context.Context, params types.GetStarGiftsParams) (*types.GetStarGiftsResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	result, err := c.API.PaymentsGetStarGifts(ctx, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get star gifts: %w", err)
	}

	gifts, ok := result.AsModified()
	if !ok {
		return &types.GetStarGiftsResult{Gifts: []types.GiftItem{}, Count: 0}, nil
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	var items []types.GiftItem
	for _, g := range gifts.Gifts {
		if len(items) >= limit {
			break
		}
		switch gift := g.(type) {
		case *tg.StarGift:
			title := gift.Title
			if title == "" {
				title = giftNames[gift.ID]
			}
			items = append(items, types.GiftItem{
				ID:                  gift.ID,
				Stars:               gift.Stars,
				ConvertStars:        gift.ConvertStars,
				Limited:             gift.Limited,
				SoldOut:             gift.SoldOut,
				Birthday:            gift.Birthday,
				AvailabilityRemains: gift.AvailabilityRemains,
				AvailabilityTotal:   gift.AvailabilityTotal,
				Title:               title,
				UpgradeStars:        gift.UpgradeStars,
				Slug:                gift.AuctionSlug,
			})
		case *tg.StarGiftUnique:
			items = append(items, types.GiftItem{
				ID:    gift.ID,
				Title: gift.Title,
				Slug:  gift.Slug,
			})
		}
	}

	return &types.GetStarGiftsResult{
		Gifts: items,
		Count: len(items),
	}, nil
}

// SendStarGift buys a star gift from the catalog and sends it to a peer.
func (c *Client) SendStarGift(ctx context.Context, params types.SendStarGiftParams) (*types.SendStarGiftResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	// Resolve gift name to ID if needed
	if params.GiftID == 0 && params.Name != "" {
		params.GiftID, err = c.resolveGiftByName(ctx, params.Name)
		if err != nil {
			return nil, err
		}
	}

	// Create invoice for the catalog gift purchase
	invoice := &tg.InputInvoiceStarGift{
		Peer:   inputPeer,
		GiftID: params.GiftID,
	}
	if params.HideName {
		invoice.SetHideName(true)
	}
	if params.Message != "" {
		invoice.SetMessage(tg.TextWithEntities{Text: params.Message})
	}

	// Get payment form
	paymentForm, err := c.API.PaymentsGetPaymentForm(ctx, &tg.PaymentsGetPaymentFormRequest{
		Invoice: invoice,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get payment form for gift: %w", err)
	}

	// Extract form ID
	starGiftForm, ok := paymentForm.(*tg.PaymentsPaymentFormStarGift)
	if !ok {
		return nil, fmt.Errorf("unexpected payment form type: %T", paymentForm)
	}

	// Send stars payment
	_, err = c.API.PaymentsSendStarsForm(ctx, &tg.PaymentsSendStarsFormRequest{
		FormID:  starGiftForm.FormID,
		Invoice: invoice,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to pay for gift: %w", err)
	}

	return &types.SendStarGiftResult{Success: true}, nil
}

// GetSavedGifts returns saved star gifts for a peer (defaults to self).
//
//nolint:funlen,lll // Maps many Telegram gift fields to result struct
func (c *Client) GetSavedGifts(ctx context.Context, params types.GetSavedGiftsParams) (*types.GetSavedGiftsResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	var inputPeer tg.InputPeerClass
	if params.Peer == "" {
		inputPeer = &tg.InputPeerSelf{}
	} else {
		var err error
		inputPeer, err = c.ResolvePeer(ctx, params.Peer)
		if err != nil {
			return nil, err
		}
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	result, err := c.API.PaymentsGetSavedStarGifts(ctx, &tg.PaymentsGetSavedStarGiftsRequest{
		Peer:   inputPeer,
		Offset: params.Offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get saved gifts: %w", err)
	}

	var items []types.SavedGiftItem
	for _, sg := range result.Gifts {
		item := types.SavedGiftItem{
			MsgID:      sg.MsgID,
			SavedID:    sg.SavedID,
			Date:       sg.Date,
			NameHidden: sg.NameHidden,
			Unsaved:    sg.Unsaved,
			CanUpgrade: sg.CanUpgrade,
		}

		if sg.FromID != nil {
			item.FromID = fmt.Sprintf("%v", sg.FromID)
		}
		if sg.Message.Text != "" {
			item.Message = sg.Message.Text
		}
		item.ConvertStars = sg.ConvertStars
		item.TransferStars = sg.TransferStars

		switch gift := sg.Gift.(type) {
		case *tg.StarGift:
			item.GiftID = gift.ID
			item.Stars = gift.Stars
			item.Title = gift.Title
			item.Slug = gift.AuctionSlug
		case *tg.StarGiftUnique:
			item.GiftID = gift.ID
			item.Title = gift.Title
			item.Slug = gift.Slug
			if len(gift.ResellAmount) > 0 {
				item.ResellStars = gift.ResellAmount[0].GetAmount()
			}
		}

		items = append(items, item)
	}

	return &types.GetSavedGiftsResult{
		Gifts:      items,
		Count:      result.Count,
		NextOffset: result.NextOffset,
	}, nil
}

// TransferStarGift transfers a star gift to another peer using the star payment flow.
//
//nolint:lll // Long function signature due to Go generics pattern
func (c *Client) TransferStarGift(ctx context.Context, params types.TransferStarGiftParams) (*types.TransferStarGiftResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	giftInput := resolveGiftInput(params.MsgID, params.Slug)

	// Create invoice for the star gift transfer
	invoice := &tg.InputInvoiceStarGiftTransfer{
		Stargift: giftInput,
		ToID:     inputPeer,
	}

	// Get payment form
	paymentForm, err := c.API.PaymentsGetPaymentForm(ctx, &tg.PaymentsGetPaymentFormRequest{
		Invoice: invoice,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get payment form for transfer: %w", err)
	}

	// Extract form ID from the star gift payment form
	starGiftForm, ok := paymentForm.(*tg.PaymentsPaymentFormStarGift)
	if !ok {
		return nil, fmt.Errorf("unexpected payment form type: %T", paymentForm)
	}

	// Send stars payment to complete the transfer
	_, err = c.API.PaymentsSendStarsForm(ctx, &tg.PaymentsSendStarsFormRequest{
		FormID:  starGiftForm.FormID,
		Invoice: invoice,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to pay for gift transfer: %w", err)
	}

	return &types.TransferStarGiftResult{Success: true}, nil
}

// BuyResaleGift buys a gift from the marketplace using the star payment flow.
//
//nolint:lll // Long function signature due to Go generics pattern
func (c *Client) BuyResaleGift(ctx context.Context, params types.BuyResaleGiftParams) (*types.BuyResaleGiftResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	// Default to self if no peer specified
	var inputPeer tg.InputPeerClass
	if params.Peer == "" {
		inputPeer = &tg.InputPeerSelf{}
	} else {
		var err error
		inputPeer, err = c.ResolvePeer(ctx, params.Peer)
		if err != nil {
			return nil, err
		}
	}

	// Create invoice for the resale gift purchase
	invoice := &tg.InputInvoiceStarGiftResale{
		Slug: params.Slug,
		ToID: inputPeer,
	}

	// Get payment form
	paymentForm, err := c.API.PaymentsGetPaymentForm(ctx, &tg.PaymentsGetPaymentFormRequest{
		Invoice: invoice,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get payment form for purchase: %w", err)
	}

	// Extract form ID from the star gift payment form
	starGiftForm, ok := paymentForm.(*tg.PaymentsPaymentFormStarGift)
	if !ok {
		return nil, fmt.Errorf("unexpected payment form type: %T", paymentForm)
	}

	// Send stars payment to complete the purchase
	_, err = c.API.PaymentsSendStarsForm(ctx, &tg.PaymentsSendStarsFormRequest{
		FormID:  starGiftForm.FormID,
		Invoice: invoice,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to pay for gift purchase: %w", err)
	}

	return &types.BuyResaleGiftResult{Success: true}, nil
}

// ConvertStarGift converts a star gift to stars.
//
//nolint:lll // Long function signature due to Go generics pattern
func (c *Client) ConvertStarGift(ctx context.Context, params types.ConvertStarGiftParams) (*types.ConvertStarGiftResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	giftInput := resolveGiftInput(params.MsgID, params.Slug)

	_, err := c.API.PaymentsConvertStarGift(ctx, giftInput)
	if err != nil {
		return nil, fmt.Errorf("failed to convert star gift: %w", err)
	}

	return &types.ConvertStarGiftResult{Success: true}, nil
}

// UpdateGiftPrice sets or updates the resale price on a star gift.
//
//nolint:lll // Long function signature due to Go generics pattern
func (c *Client) UpdateGiftPrice(ctx context.Context, params types.UpdateGiftPriceParams) (*types.UpdateGiftPriceResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	giftInput := resolveGiftInput(params.MsgID, params.Slug)

	_, err := c.API.PaymentsUpdateStarGiftPrice(ctx, &tg.PaymentsUpdateStarGiftPriceRequest{
		Stargift:     giftInput,
		ResellAmount: &tg.StarsAmount{Amount: params.Price},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update gift price: %w", err)
	}

	return &types.UpdateGiftPriceResult{Success: true}, nil
}

// GetBalance returns the current stars and TON balance.
func (c *Client) GetBalance(ctx context.Context, _ types.GetBalanceParams) (*types.GetBalanceResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	result := &types.GetBalanceResult{}

	// Get stars balance
	starsStatus, err := c.API.PaymentsGetStarsStatus(ctx, &tg.PaymentsGetStarsStatusRequest{
		Peer: &tg.InputPeerSelf{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get stars balance: %w", err)
	}
	if starsStatus.Balance != nil {
		result.Stars = starsStatus.Balance.GetAmount()
		if sa, ok := starsStatus.Balance.(*tg.StarsAmount); ok {
			result.Nanos = sa.Nanos
		}
	}

	// Get TON balance
	tonStatus, err := c.API.PaymentsGetStarsStatus(ctx, &tg.PaymentsGetStarsStatusRequest{
		Peer: &tg.InputPeerSelf{},
		Ton:  true,
	})
	if err == nil && tonStatus.Balance != nil {
		result.Ton = tonStatus.Balance.GetAmount()
	}

	return result, nil
}

// OfferGift makes an offer to buy someone's gift.
func (c *Client) OfferGift(ctx context.Context, params types.OfferGiftParams) (*types.OfferGiftResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = c.API.PaymentsSendStarGiftOffer(ctx, &tg.PaymentsSendStarGiftOfferRequest{
		Peer:     inputPeer,
		Slug:     params.Slug,
		Price:    &tg.StarsAmount{Amount: params.Price},
		Duration: params.Duration,
		RandomID: rand.Int63(), //nolint:gosec // non-crypto random is fine for RandomID
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send gift offer: %w", err)
	}

	return &types.OfferGiftResult{Success: true}, nil
}

// AcceptGiftOffer accepts an incoming gift offer.
//
//nolint:lll // Long function signature due to Go generics pattern
func (c *Client) AcceptGiftOffer(ctx context.Context, params types.AcceptGiftOfferParams) (*types.AcceptGiftOfferResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	_, err := c.API.PaymentsResolveStarGiftOffer(ctx, &tg.PaymentsResolveStarGiftOfferRequest{
		OfferMsgID: params.OfferMsgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to accept gift offer: %w", err)
	}

	return &types.AcceptGiftOfferResult{Success: true}, nil
}

// DeclineGiftOffer declines an incoming gift offer.
//
//nolint:lll // Long function signature due to Go generics pattern
func (c *Client) DeclineGiftOffer(ctx context.Context, params types.DeclineGiftOfferParams) (*types.DeclineGiftOfferResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	_, err := c.API.PaymentsResolveStarGiftOffer(ctx, &tg.PaymentsResolveStarGiftOfferRequest{
		Decline:    true,
		OfferMsgID: params.OfferMsgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to decline gift offer: %w", err)
	}

	return &types.DeclineGiftOfferResult{Success: true}, nil
}

// GetGiftAttrs returns all available attributes (models, patterns, backdrops) for a gift type.
func (c *Client) GetGiftAttrs(ctx context.Context, params types.GetGiftAttrsParams) (*types.GetGiftAttrsResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	giftID := params.GiftID
	if giftID == 0 && params.Name != "" {
		var err error
		giftID, err = c.resolveGiftByName(ctx, params.Name)
		if err != nil {
			return nil, err
		}
	}

	attrResult, err := c.API.PaymentsGetResaleStarGifts(ctx, &tg.PaymentsGetResaleStarGiftsRequest{
		GiftID: giftID,
		Limit:  1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get gift attributes: %w", err)
	}

	result := &types.GetGiftAttrsResult{
		GiftID:    giftID,
		Models:    []types.GiftAttribute{},
		Patterns:  []types.GiftAttribute{},
		Backdrops: []types.GiftAttribute{},
	}

	attrs, _ := attrResult.GetAttributes()
	for _, attr := range attrs {
		switch a := attr.(type) {
		case *tg.StarGiftAttributeModel:
			result.Models = append(result.Models, types.GiftAttribute{
				Type: "model", Name: a.Name, RarityPermille: a.RarityPermille,
			})
		case *tg.StarGiftAttributePattern:
			result.Patterns = append(result.Patterns, types.GiftAttribute{
				Type: "pattern", Name: a.Name, RarityPermille: a.RarityPermille,
			})
		case *tg.StarGiftAttributeBackdrop:
			result.Backdrops = append(result.Backdrops, types.GiftAttribute{
				Type: "backdrop", Name: a.Name, RarityPermille: a.RarityPermille,
			})
		}
	}

	result.Count = len(result.Models) + len(result.Patterns) + len(result.Backdrops)
	return result, nil
}

// resolveAttributeFilters resolves attribute name filters to IDs by fetching available attributes.
//
//nolint:lll // Long function signature due to multiple filter parameters
func (c *Client) resolveAttributeFilters(ctx context.Context, giftID int64, model, pattern, backdrop string) ([]tg.StarGiftAttributeIDClass, error) {
	attrResult, err := c.API.PaymentsGetResaleStarGifts(ctx, &tg.PaymentsGetResaleStarGiftsRequest{
		GiftID: giftID,
		Limit:  1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get available attributes: %w", err)
	}

	var filters []tg.StarGiftAttributeIDClass
	attrs, _ := attrResult.GetAttributes()
	for _, attr := range attrs {
		switch a := attr.(type) {
		case *tg.StarGiftAttributeModel:
			if model != "" && strings.EqualFold(a.Name, model) {
				if doc, ok := a.Document.(*tg.Document); ok {
					filters = append(filters, &tg.StarGiftAttributeIDModel{DocumentID: doc.ID})
				}
			}
		case *tg.StarGiftAttributePattern:
			if pattern != "" && strings.EqualFold(a.Name, pattern) {
				if doc, ok := a.Document.(*tg.Document); ok {
					filters = append(filters, &tg.StarGiftAttributeIDPattern{DocumentID: doc.ID})
				}
			}
		case *tg.StarGiftAttributeBackdrop:
			if backdrop != "" && strings.EqualFold(a.Name, backdrop) {
				filters = append(filters, &tg.StarGiftAttributeIDBackdrop{BackdropID: a.BackdropID})
			}
		}
	}
	return filters, nil
}

//nolint:funlen,lll // Maps many Telegram resale gift fields to result struct
// GetResaleGifts returns gifts listed for resale for a given gift type.
func (c *Client) GetResaleGifts(ctx context.Context, params types.GetResaleGiftsParams) (*types.GetResaleGiftsResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	// Resolve gift name to ID if needed
	if params.GiftID == 0 && params.Name != "" {
		var err error
		params.GiftID, err = c.resolveGiftByName(ctx, params.Name)
		if err != nil {
			return nil, err
		}
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}

	req := &tg.PaymentsGetResaleStarGiftsRequest{
		GiftID: params.GiftID,
		Offset: params.Offset,
		Limit:  limit,
	}
	if params.SortByPrice {
		req.SetSortByPrice(true)
	}
	if params.SortByNum {
		req.SetSortByNum(true)
	}

	// Resolve attribute name filters to IDs
	hasFilters := params.Model != "" || params.Pattern != "" || params.Backdrop != ""
	if hasFilters {
		filters, err := c.resolveAttributeFilters(ctx, params.GiftID, params.Model, params.Pattern, params.Backdrop)
		if err != nil {
			return nil, err
		}
		if len(filters) > 0 {
			req.SetAttributes(filters)
		}
	}

	result, err := c.API.PaymentsGetResaleStarGifts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get resale gifts: %w", err)
	}

	var items []types.ResaleGiftItem
	for _, g := range result.Gifts {
		gift, ok := g.(*tg.StarGiftUnique)
		if !ok {
			continue
		}
		item := types.ResaleGiftItem{
			ID:                 gift.ID,
			GiftID:             gift.GiftID,
			Title:              gift.Title,
			Slug:               gift.Slug,
			Num:                gift.Num,
			OwnerName:          gift.OwnerName,
			AvailabilityIssued: gift.AvailabilityIssued,
			AvailabilityTotal:  gift.AvailabilityTotal,
		}
		if len(gift.ResellAmount) > 0 {
			item.ResellStars = gift.ResellAmount[0].GetAmount()
		}
		for _, attr := range gift.Attributes {
			switch a := attr.(type) {
			case *tg.StarGiftAttributeModel:
				item.Attributes = append(item.Attributes, types.GiftAttribute{
					Type: "model", Name: a.Name, RarityPermille: a.RarityPermille,
				})
			case *tg.StarGiftAttributePattern:
				item.Attributes = append(item.Attributes, types.GiftAttribute{
					Type: "pattern", Name: a.Name, RarityPermille: a.RarityPermille,
				})
			case *tg.StarGiftAttributeBackdrop:
				item.Attributes = append(item.Attributes, types.GiftAttribute{
					Type: "backdrop", Name: a.Name, RarityPermille: a.RarityPermille,
				})
			}
		}
		items = append(items, item)
	}

	res := &types.GetResaleGiftsResult{
		Gifts: items,
		Count: result.Count,
	}
	if offset, ok := result.GetNextOffset(); ok {
		res.NextOffset = offset
	}

	return res, nil
}

// GetGiftValue returns value/pricing analytics for a unique star gift.
func (c *Client) GetGiftValue(ctx context.Context, params types.GetGiftValueParams) (*types.GetGiftValueResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	result, err := c.API.PaymentsGetUniqueStarGiftValueInfo(ctx, params.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get gift value info: %w", err)
	}

	info := &types.GetGiftValueResult{
		Currency:           result.Currency,
		Value:              result.Value,
		InitialSaleDate:    result.InitialSaleDate,
		InitialSaleStars:   result.InitialSaleStars,
		InitialSalePrice:   result.InitialSalePrice,
		LastSaleOnFragment: result.LastSaleOnFragment,
		ValueIsAverage:     result.ValueIsAverage,
	}

	if v, ok := result.GetLastSaleDate(); ok {
		info.LastSaleDate = v
	}
	if v, ok := result.GetLastSalePrice(); ok {
		info.LastSalePrice = v
	}
	if v, ok := result.GetFloorPrice(); ok {
		info.FloorPrice = v
	}
	if v, ok := result.GetAveragePrice(); ok {
		info.AveragePrice = v
	}
	if v, ok := result.GetListedCount(); ok {
		info.ListedCount = v
	}
	if v, ok := result.GetFragmentListedCount(); ok {
		info.FragmentListedCount = v
	}
	if v, ok := result.GetFragmentListedURL(); ok {
		info.FragmentListedURL = v
	}

	return info, nil
}

//nolint:funlen,nestif // Maps many Telegram gift info fields to result struct
// GetGiftInfo returns detailed info about a unique star gift by slug.
func (c *Client) GetGiftInfo(ctx context.Context, params types.GetGiftInfoParams) (*types.GetGiftInfoResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	result, err := c.API.PaymentsGetUniqueStarGift(ctx, params.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get gift info: %w", err)
	}

	gift, ok := result.Gift.(*tg.StarGiftUnique)
	if !ok {
		return nil, fmt.Errorf("unexpected gift type for slug %s", params.Slug)
	}

	info := &types.GetGiftInfoResult{
		ID:                 gift.ID,
		GiftID:             gift.GiftID,
		Title:              gift.Title,
		Slug:               gift.Slug,
		Num:                gift.Num,
		OwnerName:          gift.OwnerName,
		OwnerAddress:       gift.OwnerAddress,
		AvailabilityIssued: gift.AvailabilityIssued,
		AvailabilityTotal:  gift.AvailabilityTotal,
		GiftAddress:        gift.GiftAddress,
	}
	if ownerID, ok := gift.GetOwnerID(); ok {
		if p, ok := ownerID.(*tg.PeerUser); ok {
			info.OwnerID = fmt.Sprintf("%d", p.UserID)
		}
	}
	// Fallback: find owner from Users list by matching name
	if info.OwnerID == "" && gift.OwnerName != "" {
		for _, u := range result.Users {
			if user, ok := u.(*tg.User); ok {
				name := user.FirstName
				if user.LastName != "" {
					name += " " + user.LastName
				}
				if name == gift.OwnerName {
					info.OwnerID = fmt.Sprintf("%d", user.ID)
					break
				}
			}
		}
	}

	if len(gift.ResellAmount) > 0 {
		info.ResellStars = gift.ResellAmount[0].GetAmount()
	}

	for _, attr := range gift.Attributes {
		switch a := attr.(type) {
		case *tg.StarGiftAttributeModel:
			info.Attributes = append(info.Attributes, types.GiftAttribute{
				Type:           "model",
				Name:           a.Name,
				RarityPermille: a.RarityPermille,
			})
		case *tg.StarGiftAttributePattern:
			info.Attributes = append(info.Attributes, types.GiftAttribute{
				Type:           "pattern",
				Name:           a.Name,
				RarityPermille: a.RarityPermille,
			})
		case *tg.StarGiftAttributeBackdrop:
			info.Attributes = append(info.Attributes, types.GiftAttribute{
				Type:           "backdrop",
				Name:           a.Name,
				RarityPermille: a.RarityPermille,
			})
		}
	}

	return info, nil
}
