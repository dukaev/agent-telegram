package reaction

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

	for name, call := range map[string]func() error{
		"AddReaction": func() error {
			_, err := c.AddReaction(ctx, types.AddReactionParams{})
			return err
		},
		"RemoveReaction": func() error {
			_, err := c.RemoveReaction(ctx, types.RemoveReactionParams{})
			return err
		},
		"ListReactions": func() error {
			_, err := c.ListReactions(ctx, types.ListReactionsParams{})
			return err
		},
	} {
		if err := call(); !errors.Is(err, client.ErrNotInitialized) {
			t.Fatalf("%s err = %v, want ErrNotInitialized", name, err)
		}
	}
}
