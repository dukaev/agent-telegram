package gift

import (
	"context"
	"errors"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"

	"agent-telegram/telegram/client"
	"agent-telegram/telegram/types"
)

func TestClientMethodsRequireInitialization(t *testing.T) {
	c := NewClient(nil)
	ctx := context.Background()
	check := func(name string, err error) {
		t.Helper()
		if !errors.Is(err, client.ErrNotInitialized) {
			t.Fatalf("%s err = %v, want ErrNotInitialized", name, err)
		}
	}

	_, err := c.GetStarGifts(ctx, types.GetStarGiftsParams{})
	check("GetStarGifts", err)
	_, err = c.SendStarGift(ctx, types.SendStarGiftParams{})
	check("SendStarGift", err)
	_, err = c.GetSavedGifts(ctx, types.GetSavedGiftsParams{})
	check("GetSavedGifts", err)
	_, err = c.TransferStarGift(ctx, types.TransferStarGiftParams{})
	check("TransferStarGift", err)
	_, err = c.BuyResaleGift(ctx, types.BuyResaleGiftParams{})
	check("BuyResaleGift", err)
	_, err = c.ConvertStarGift(ctx, types.ConvertStarGiftParams{})
	check("ConvertStarGift", err)
	_, err = c.UpdateGiftPrice(ctx, types.UpdateGiftPriceParams{})
	check("UpdateGiftPrice", err)
	_, err = c.GetBalance(ctx, types.GetBalanceParams{})
	check("GetBalance", err)
	_, err = c.OfferGift(ctx, types.OfferGiftParams{})
	check("OfferGift", err)
	_, err = c.AcceptGiftOffer(ctx, types.AcceptGiftOfferParams{})
	check("AcceptGiftOffer", err)
	_, err = c.DeclineGiftOffer(ctx, types.DeclineGiftOfferParams{})
	check("DeclineGiftOffer", err)
	_, err = c.GetGiftAttrs(ctx, types.GetGiftAttrsParams{})
	check("GetGiftAttrs", err)
	_, err = c.GetResaleGifts(ctx, types.GetResaleGiftsParams{})
	check("GetResaleGifts", err)
	_, err = c.GetGiftValue(ctx, types.GetGiftValueParams{})
	check("GetGiftValue", err)
	_, err = c.GetGiftInfo(ctx, types.GetGiftInfoParams{})
	check("GetGiftInfo", err)
}

func TestGiftHelpers(t *testing.T) {
	if got, err := (&Client{}).resolveGiftByName(context.Background(), "heart"); err != nil || got == 0 {
		t.Fatalf("known gift by name = %d, %v", got, err)
	}
	if input := resolveGiftInput(0, "Gift-1"); input.(*tg.InputSavedStarGiftSlug).Slug != "Gift-1" {
		t.Fatalf("slug input = %#v", input)
	}
	if input := resolveGiftInput(42, ""); input.(*tg.InputSavedStarGiftUser).MsgID != 42 {
		t.Fatalf("msg input = %#v", input)
	}
}

func TestGetStarGiftsWithFakeAPI(t *testing.T) {
	c := NewClient(nil)
	c.SetAPI(tg.NewClient(tgmock.Invoker(func(input bin.Encoder) (bin.Encoder, error) {
		switch input.(type) {
		case *tg.PaymentsGetStarGiftsRequest:
			return &tg.PaymentsStarGifts{Gifts: []tg.StarGiftClass{
				&tg.StarGift{
					ID:                  1,
					Stars:               100,
					ConvertStars:        50,
					Sticker:             &tg.DocumentEmpty{ID: 1},
					Limited:             true,
					SoldOut:             true,
					Birthday:            true,
					AvailabilityRemains: 2,
					AvailabilityTotal:   10,
					Title:               "Heart",
					UpgradeStars:        25,
					AuctionSlug:         "Heart-1",
				},
			}}, nil
		default:
			t.Fatalf("unexpected request %T", input)
			return nil, nil
		}
	})))

	result, err := c.GetStarGifts(context.Background(), types.GetStarGiftsParams{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if result.Count != 1 || result.Gifts[0].Title != "Heart" || result.Gifts[0].Slug != "Heart-1" {
		t.Fatalf("result = %+v", result)
	}
}
