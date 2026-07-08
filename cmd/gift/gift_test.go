package gift

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestAddGiftCommandRegistersExpectedSurface(t *testing.T) {
	root := &cobra.Command{Use: "root"}

	AddGiftCommand(root)

	giftCmd := childCommand(root, "gift")
	if giftCmd == nil {
		t.Fatal("gift command was not registered")
	}
	for _, name := range []string{
		"list", "send", "convert", "price", "offer", "info",
		"attrs", "accept", "decline",
	} {
		if childCommand(giftCmd, name) == nil {
			t.Fatalf("gift subcommand %q was not registered", name)
		}
	}
	if ListCmd.Flags().Lookup("sort-price") == nil || ListCmd.Flags().Lookup("model") == nil {
		t.Fatal("expected list filters")
	}
	if SendCmd.Flags().Lookup("to") == nil || SendCmd.Flags().Lookup("msg-id") == nil {
		t.Fatal("expected send flags")
	}
}

func TestGiftListPrinters(t *testing.T) {
	output := captureStderr(t, func() {
		printGiftList(map[string]any{
			"count": int64(1),
			"gifts": []any{
				map[string]any{
					"id":                  int64(1),
					"title":               "Heart",
					"stars":               int64(100),
					"limited":             true,
					"soldOut":             true,
					"birthday":            true,
					"availabilityRemains": int64(2),
					"availabilityTotal":   int64(10),
					"slug":                "Heart-1",
				},
				"ignored",
			},
		})
		printSavedGifts(map[string]any{
			"count":      int64(1),
			"nextOffset": "next",
			"gifts": []any{
				map[string]any{
					"giftId":        int64(2),
					"msgId":         int64(22),
					"title":         "Bear",
					"stars":         int64(50),
					"convertStars":  int64(10),
					"transferStars": int64(5),
					"resellStars":   int64(99),
					"fromId":        "user",
					"slug":          "Bear-2",
				},
			},
		})
		printResaleGifts(map[string]any{
			"count":      int64(1),
			"nextOffset": "more",
			"gifts": []any{
				map[string]any{
					"title":       "Watch",
					"num":         int64(7),
					"resellStars": int64(1000),
					"ownerName":   "Ada",
					"slug":        "Watch-7",
					"attributes": []any{
						map[string]any{"name": "Ruby"},
						map[string]any{},
						"ignored",
					},
				},
			},
		})
	})

	for _, want := range []string{
		"Star Gifts (1)", "Heart", "sold out", "Saved Gifts (1)",
		"Summary: 1 gifts", "Resale Listings (1)", "Watch", "Ruby",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}

	fallback := captureStderr(t, func() {
		printGiftList("bad")
		printSavedGifts("bad")
		printResaleGifts("bad")
	})
	for _, want := range []string{"Failed to get gifts", "Failed to get saved gifts", "Failed to get resale gifts"} {
		if !strings.Contains(fallback, want) {
			t.Fatalf("fallback missing %q:\n%s", want, fallback)
		}
	}
}

func TestGiftInfoValueAndAttrsPrinters(t *testing.T) {
	output := captureStderr(t, func() {
		PrintGiftInfo(map[string]any{
			"title":              "Watch",
			"slug":               "Watch-7",
			"num":                int64(7),
			"giftId":             int64(77),
			"availabilityIssued": int64(10),
			"availabilityTotal":  int64(100),
			"ownerName":          "Ada",
			"ownerId":            "42",
			"ownerAddress":       "owner-address",
			"giftAddress":        "gift-address",
			"resellStars":        int64(500),
			"attributes": []any{
				map[string]any{"type": "model", "name": "Ruby", "rarityPermille": float64(25)},
				"ignored",
			},
		})
		printGiftValue(map[string]any{
			"currency":            "USD",
			"value":               int64(100),
			"valueIsAverage":      true,
			"initialSaleStars":    int64(10),
			"initialSalePrice":    int64(20),
			"initialSaleDate":     int64(1704067200),
			"floorPrice":          int64(90),
			"averagePrice":        int64(110),
			"lastSalePrice":       int64(120),
			"lastSaleDate":        int64(1704153600),
			"lastSaleOnFragment":  true,
			"listedCount":         int64(3),
			"fragmentListedCount": int64(2),
			"fragmentListedUrl":   "https://fragment.example",
		})
		printGiftAttrs(map[string]any{
			"models":    []any{map[string]any{"name": "Ruby", "rarityPermille": float64(25)}},
			"patterns":  []any{map[string]any{"name": "Wave"}},
			"backdrops": []any{map[string]any{"name": "Blue"}},
		}, "")
		printGiftInfoHints("Watch-7")
	})

	for _, want := range []string{
		"Watch [Watch-7]", "Gift ID: 77", "Owner: Ada", "Attributes:",
		"Estimated value (avg): 100 USD", "Initial sale: 10 stars",
		"Last sale: 120 stars", "Models (1)", "Patterns (1)",
		"Related commands:",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}

	fallback := captureStderr(t, func() {
		PrintGiftInfo("bad")
		printGiftAttrs("bad", "")
		printGiftValue("bad")
	})
	for _, want := range []string{"Failed to get gift info", "Failed to get gift attributes"} {
		if !strings.Contains(fallback, want) {
			t.Fatalf("fallback missing %q:\n%s", want, fallback)
		}
	}
}

func childCommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, child := range cmd.Commands() {
		if child.Name() == name {
			return child
		}
	}
	return nil
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	defer func() { os.Stderr = old }()

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return <-done
}
