// Package main provides a utility to sanitize recorded fixtures.
//
// Usage:
//
//	go run ./testdata/sanitizer -input raw_fixture.json -output sanitized.json
//
// Or to sanitize in place:
//
//	go run ./testdata/sanitizer -input fixture.json -inplace
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	input   = flag.String("input", "", "Input fixture file")
	output  = flag.String("output", "", "Output file (default: stdout)")
	inplace = flag.Bool("inplace", false, "Modify file in place")
)

// Mappings for sanitization
var (
	userIDMap    = make(map[int64]int64)
	chatIDMap    = make(map[int64]int64)
	channelIDMap = make(map[int64]int64)
	hashMap      = make(map[int64]int64)
	usernameMap  = make(map[string]string)
	phoneMap     = make(map[string]string)

	nextUserID    int64 = 100000001
	nextChatID    int64 = 200000001
	nextChannelID int64 = 300000001
	nextHash      int64 = 1000000000000000001
	nextUsername        = 1
	nextPhone           = 1

	testNames = []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank"}
	nameIndex = 0
)

func main() {
	flag.Parse()

	if *input == "" {
		log.Fatal("input file is required")
	}

	data, err := os.ReadFile(*input)
	if err != nil {
		log.Fatalf("failed to read input: %v", err)
	}

	var fixture map[string]any
	if err := json.Unmarshal(data, &fixture); err != nil {
		log.Fatalf("failed to parse JSON: %v", err)
	}

	// Sanitize the fixture
	sanitized := sanitizeValue(fixture)

	// Mark as sanitized in meta
	if meta, ok := sanitized.(map[string]any)["meta"].(map[string]any); ok {
		meta["sanitized"] = true
	}

	// Output
	result, err := json.MarshalIndent(sanitized, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal: %v", err)
	}

	if *inplace {
		if err := os.WriteFile(*input, result, 0o644); err != nil {
			log.Fatalf("failed to write: %v", err)
		}
		fmt.Printf("Sanitized %s in place\n", *input)
	} else if *output != "" {
		if err := os.WriteFile(*output, result, 0o644); err != nil {
			log.Fatalf("failed to write: %v", err)
		}
		fmt.Printf("Sanitized fixture written to %s\n", *output)
	} else {
		fmt.Println(string(result))
	}

	// Print mapping summary
	fmt.Fprintln(os.Stderr, "\n--- Sanitization Mapping ---")
	fmt.Fprintf(os.Stderr, "Users:    %d mapped\n", len(userIDMap))
	fmt.Fprintf(os.Stderr, "Chats:    %d mapped\n", len(chatIDMap))
	fmt.Fprintf(os.Stderr, "Channels: %d mapped\n", len(channelIDMap))
	fmt.Fprintf(os.Stderr, "Hashes:   %d mapped\n", len(hashMap))
	fmt.Fprintf(os.Stderr, "Usernames: %d mapped\n", len(usernameMap))
}

func sanitizeValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return sanitizeObject(val)
	case []any:
		result := make([]any, len(val))
		for i, item := range val {
			result[i] = sanitizeValue(item)
		}
		return result
	default:
		return v
	}
}

func sanitizeObject(obj map[string]any) map[string]any {
	result := make(map[string]any)

	for key, value := range obj {
		result[key] = sanitizeField(key, value, obj)
	}

	return result
}

func sanitizeField(key string, value any, parent map[string]any) any {
	switch key {
	case "user_id":
		if id, ok := toInt64(value); ok {
			return sanitizeUserID(id)
		}
	case "chat_id":
		if id, ok := toInt64(value); ok {
			return sanitizeChatID(id)
		}
	case "channel_id":
		if id, ok := toInt64(value); ok {
			return sanitizeChannelID(id)
		}
	case "id":
		// Check context from parent's "_" field
		if objType, ok := parent["_"].(string); ok {
			if id, ok := toInt64(value); ok {
				switch {
				case strings.Contains(objType, "user"):
					return sanitizeUserID(id)
				case strings.Contains(objType, "chat") && !strings.Contains(objType, "channel"):
					return sanitizeChatID(id)
				case strings.Contains(objType, "channel"):
					return sanitizeChannelID(id)
				}
			}
		}
	case "access_hash":
		if hash, ok := toInt64(value); ok {
			return sanitizeHash(hash)
		}
	case "username":
		if username, ok := value.(string); ok {
			return sanitizeUsername(username)
		}
	case "phone":
		if phone, ok := value.(string); ok {
			return sanitizePhone(phone)
		}
	case "first_name":
		return getTestFirstName()
	case "last_name":
		return "Test"
	case "message":
		if msg, ok := value.(string); ok {
			return sanitizeMessage(msg)
		}
	case "title":
		if title, ok := value.(string); ok {
			return "Test " + sanitizeTitle(title)
		}
	}

	// Recurse into nested values
	return sanitizeValue(value)
}

func toInt64(v any) (int64, bool) {
	switch val := v.(type) {
	case float64:
		return int64(val), true
	case int64:
		return val, true
	case int:
		return int64(val), true
	case json.Number:
		i, err := val.Int64()
		return i, err == nil
	}
	return 0, false
}

func sanitizeUserID(id int64) int64 {
	if mapped, ok := userIDMap[id]; ok {
		return mapped
	}
	mapped := nextUserID
	nextUserID++
	userIDMap[id] = mapped
	return mapped
}

func sanitizeChatID(id int64) int64 {
	if mapped, ok := chatIDMap[id]; ok {
		return mapped
	}
	mapped := nextChatID
	nextChatID++
	chatIDMap[id] = mapped
	return mapped
}

func sanitizeChannelID(id int64) int64 {
	if mapped, ok := channelIDMap[id]; ok {
		return mapped
	}
	mapped := nextChannelID
	nextChannelID++
	channelIDMap[id] = mapped
	return mapped
}

func sanitizeHash(hash int64) int64 {
	if mapped, ok := hashMap[hash]; ok {
		return mapped
	}
	mapped := nextHash
	nextHash++
	hashMap[hash] = mapped
	return mapped
}

func sanitizeUsername(username string) string {
	if mapped, ok := usernameMap[username]; ok {
		return mapped
	}
	mapped := "testuser" + strconv.Itoa(nextUsername)
	nextUsername++
	usernameMap[username] = mapped
	return mapped
}

func sanitizePhone(phone string) string {
	if mapped, ok := phoneMap[phone]; ok {
		return mapped
	}
	mapped := fmt.Sprintf("+1555123%04d", nextPhone)
	nextPhone++
	phoneMap[phone] = mapped
	return mapped
}

func sanitizeMessage(msg string) string {
	// Keep structure but replace content
	if len(msg) > 50 {
		return "Test message (long content sanitized)"
	}
	return "Test message"
}

func sanitizeTitle(title string) string {
	// Remove any identifying words but keep type hints
	title = regexp.MustCompile(`[A-Z][a-z]+`).ReplaceAllString(title, "")
	if strings.Contains(strings.ToLower(title), "channel") {
		return "Channel"
	}
	if strings.Contains(strings.ToLower(title), "group") {
		return "Group"
	}
	return "Chat"
}

func getTestFirstName() string {
	name := testNames[nameIndex%len(testNames)]
	nameIndex++
	return name
}
