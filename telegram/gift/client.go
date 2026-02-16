// Package gift provides Telegram star gift operations.
package gift

import (
	"context"
	"fmt"
	"math/rand"

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
			items = append(items, types.GiftItem{
				ID:                  gift.ID,
				Stars:               gift.Stars,
				ConvertStars:        gift.ConvertStars,
				Limited:             gift.Limited,
				SoldOut:             gift.SoldOut,
				Birthday:            gift.Birthday,
				AvailabilityRemains: gift.AvailabilityRemains,
				AvailabilityTotal:   gift.AvailabilityTotal,
				Title:               gift.Title,
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

// SendStarGift sends a star gift to a peer.
func (c *Client) SendStarGift(ctx context.Context, params types.SendStarGiftParams) (*types.SendStarGiftResult, error) {
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
		return nil, fmt.Errorf("failed to send star gift: %w", err)
	}

	return &types.SendStarGiftResult{Success: true}, nil
}

// GetSavedGifts returns saved star gifts for a peer (defaults to self).
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
		}

		items = append(items, item)
	}

	return &types.GetSavedGiftsResult{
		Gifts:      items,
		Count:      result.Count,
		NextOffset: result.NextOffset,
	}, nil
}

// TransferStarGift transfers a star gift to another peer.
func (c *Client) TransferStarGift(ctx context.Context, params types.TransferStarGiftParams) (*types.TransferStarGiftResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	giftInput := resolveGiftInput(params.MsgID, params.Slug)

	_, err = c.API.PaymentsTransferStarGift(ctx, &tg.PaymentsTransferStarGiftRequest{
		Stargift: giftInput,
		ToID:     inputPeer,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to transfer star gift: %w", err)
	}

	return &types.TransferStarGiftResult{Success: true}, nil
}

// ConvertStarGift converts a star gift to stars.
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
