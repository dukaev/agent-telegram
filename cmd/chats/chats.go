// Package chats provides the chats command implementation.
package chats

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"agent-telegram/internal/ipc"
)

// Run executes the chats command.
func Run(args []string) {
	chatsCmd := flag.NewFlagSet("chats", flag.ExitOnError)
	socketPath := chatsCmd.String("socket", ipc.DefaultSocketPath(), "Path to Unix socket")
	jsonOutput := chatsCmd.Bool("json", false, "Output as JSON")
	limit := chatsCmd.Int("limit", 10, "Number of chats to return (max 100)")
	offset := chatsCmd.Int("offset", 0, "Offset for pagination")

	if err := chatsCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Validate limit
	if *limit < 1 {
		*limit = 1
	}
	if *limit > 100 {
		*limit = 100
	}
	if *offset < 0 {
		*offset = 0
	}

	client := ipc.NewClient(*socketPath)

	// Build params
	params := map[string]int{
		"limit":  *limit,
		"offset": *offset,
	}

	result, err := client.Call("get_chats", params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	} else {
		r, ok := result.(map[string]interface{})
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
			os.Exit(1)
		}

		chats, _ := r["chats"].([]interface{})
		limit := r["limit"].(float64)
		offset := r["offset"].(float64)
		count := r["count"].(float64)

		fmt.Printf("Chats (limit: %.0f, offset: %.0f, count: %.0f):\n", limit, offset, count)
		fmt.Println()

		for _, chat := range chats {
			chatInfo, ok := chat.(map[string]interface{})
			if !ok {
				continue
			}

			// Display based on type
			if chatType, ok := chatInfo["type"].(string); ok {
				switch chatType {
				case "user":
					name := ""
					if fn, ok := chatInfo["first_name"].(string); ok {
						name = fn
						if ln, ok := chatInfo["last_name"].(string); ok && ln != "" {
							name += " " + ln
						}
					}
					if name == "" {
						if un, ok := chatInfo["username"].(string); ok {
							name = "@" + un
						}
					}
					fmt.Printf("  👤 %s", name)
					if _, ok := chatInfo["bot"].(bool); ok {
						fmt.Print(" 🤖")
					}
					fmt.Println()

				case "chat":
					if title, ok := chatInfo["title"].(string); ok {
						fmt.Printf("  👥 %s", title)
						if count, ok := chatInfo["participants_count"].(int); ok {
							fmt.Printf(" (%d members)", count)
						}
						fmt.Println()
					}

				case "channel":
					if title, ok := chatInfo["title"].(string); ok {
						fmt.Printf("  📢 %s", title)
						if username, ok := chatInfo["username"].(string); ok && username != "" {
							fmt.Printf(" (@%s)", username)
						}
						if megagroup, ok := chatInfo["megagroup"].(bool); ok && megagroup {
							fmt.Print(" (group)")
						}
						fmt.Println()
					}
				}
			}

			// Show unread count
			if unread, ok := chatInfo["unread_count"].(int); ok && unread > 0 {
				fmt.Printf("     🔔 %d unread\n", unread)
			}
		}
	}
}
