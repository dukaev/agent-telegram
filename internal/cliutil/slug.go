package cliutil

import "strings"

// nftURLPrefixes are the known URL prefixes for Telegram NFT gift links.
var nftURLPrefixes = []string{
	"https://t.me/nft/",
	"http://t.me/nft/",
	"t.me/nft/",
}

// ParseGiftSlug extracts a gift slug from a Telegram NFT URL or returns the input as-is.
// Supported formats:
//   - https://t.me/nft/DeskCalendar-34796 → DeskCalendar-34796
//   - DeskCalendar-34796 → DeskCalendar-34796
func ParseGiftSlug(arg string) string {
	arg = strings.TrimSpace(arg)
	for _, prefix := range nftURLPrefixes {
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix)
		}
	}
	return arg
}
