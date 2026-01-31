// Package getme provides the get-me command implementation.
package getme

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"agent-telegram/internal/ipc"
)

// Run executes the get-me command.
func Run(args []string) {
	getMeCmd := flag.NewFlagSet("get-me", flag.ExitOnError)
	socketPath := getMeCmd.String("socket", ipc.DefaultSocketPath(), "Path to Unix socket")
	jsonOutput := getMeCmd.Bool("json", false, "Output as JSON")

	if err := getMeCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	client := ipc.NewClient(*socketPath)
	result, err := client.Call("get_me", nil)
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
		user, ok := result.(map[string]interface{})
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
			os.Exit(1)
		}
		fmt.Printf("User Info:\n")
		if firstName, ok := user["first_name"].(string); ok && firstName != "" {
			fmt.Printf("  Name: %s", firstName)
			if lastName, ok := user["last_name"].(string); ok && lastName != "" {
				fmt.Printf(" %s", lastName)
			}
			fmt.Println()
		}
		if username, ok := user["username"].(string); ok && username != "" {
			fmt.Printf("  Username: @%s\n", username)
		}
		if phone, ok := user["phone"].(string); ok && phone != "" {
			fmt.Printf("  Phone: %s\n", phone)
		}
		if id, ok := user["id"].(float64); ok {
			fmt.Printf("  ID: %d\n", int64(id))
		}
	}
}
