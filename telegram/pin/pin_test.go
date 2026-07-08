package pin

import (
	"context"
	"errors"
	"testing"

	"agent-telegram/telegram/client"
	"agent-telegram/telegram/types"
)

func TestClientMethodsRequireInitialization(t *testing.T) {
	c := NewClient(nil)
	ctx := context.Background()

	_, err := c.PinMessage(ctx, types.PinMessageParams{})
	if !errors.Is(err, client.ErrNotInitialized) {
		t.Fatalf("PinMessage err = %v", err)
	}
	_, err = c.UnpinMessage(ctx, types.UnpinMessageParams{})
	if !errors.Is(err, client.ErrNotInitialized) {
		t.Fatalf("UnpinMessage err = %v", err)
	}
}
