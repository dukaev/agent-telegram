package helpers

import (
	"regexp"
	"strconv"
	"unicode/utf16"

	"github.com/gotd/td/tg"
)

// customEmojiPlaceholder is a single BMP character used to replace <custom:ID> markers.
// U+2753 BLACK QUESTION MARK ORNAMENT — visible fallback, 1 UTF-16 code unit.
const customEmojiPlaceholder = "\u2753"

// customEmojiPattern matches <custom:documentId> markers in text.
var customEmojiPattern = regexp.MustCompile(`<custom:(\d+)>`)

// ParseCustomEmojis scans text for <custom:documentId> markers, replaces each
// with a placeholder character, and returns the cleaned text plus
// MessageEntityCustomEmoji entities with correct UTF-16 offsets.
// If no markers are found, returns the original text and nil.
func ParseCustomEmojis(text string) (string, []tg.MessageEntityClass) {
	matches := customEmojiPattern.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return text, nil
	}

	var (
		result   []byte
		entities []tg.MessageEntityClass
		lastEnd  int
	)

	for _, match := range matches {
		// Append text before this match
		result = append(result, text[lastEnd:match[0]]...)

		// UTF-16 offset of the placeholder position
		offset := utf16Len(string(result))

		// Parse document ID
		docID, err := strconv.ParseInt(text[match[2]:match[3]], 10, 64)
		if err != nil {
			// Invalid ID — keep original text
			result = append(result, text[match[0]:match[1]]...)
			lastEnd = match[1]
			continue
		}

		// Add placeholder
		result = append(result, customEmojiPlaceholder...)

		entities = append(entities, &tg.MessageEntityCustomEmoji{
			Offset:     offset,
			Length:     1,
			DocumentID: docID,
		})

		lastEnd = match[1]
	}

	result = append(result, text[lastEnd:]...)

	return string(result), entities
}

// utf16Len returns the number of UTF-16 code units needed to encode the string.
func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		n += utf16.RuneLen(r)
	}
	return n
}
