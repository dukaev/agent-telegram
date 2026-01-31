// Package get-updates provides the get-updates command implementation.
package getupdates

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"agent-telegram/internal/ipc"
)

// Run executes the get-updates command.
func Run(args []string) {
	getUpdatesCmd := flag.NewFlagSet("get-updates", flag.ExitOnError)
	socketPath := getUpdatesCmd.String("socket", ipc.DefaultSocketPath(), "Path to Unix socket")
	limit := getUpdatesCmd.Int("limit", 10, "Number of updates to get")
	jsonOutput := getUpdatesCmd.Bool("json", false, "Output as JSON")

	if err := getUpdatesCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	client := ipc.NewClient(*socketPath)

	params := map[string]interface{}{
		"limit": *limit,
	}

	result, err := client.Call("get_updates", params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
		return
	}

	// Pretty print
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response\n")
		os.Exit(1)
	}

	updates, ok := resultMap["updates"].([]interface{})
	if !ok {
		fmt.Println("No updates")
		return
	}

	count := len(updates)
	if count == 0 {
		fmt.Println("No updates")
		return
	}

	fmt.Printf("Updates (%d):\n", count)
	for _, u := range updates {
		update, ok := u.(map[string]interface{})
		if !ok {
			continue
		}
		fmt.Printf("  [%d] %s", update["id"], update["type"])
		if data, ok := update["data"].(map[string]interface{}); ok {
			if msg, ok := data["message"].(map[string]interface{}); ok {
				if text, ok := msg["text"].(string); ok {
					fmt.Printf(": %s", text)
				}
			}
		}
		fmt.Println()
	}
}
