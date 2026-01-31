// Package status provides the status command implementation.
package status

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"agent-telegram/internal/ipc"
)

// Run executes the status command.
func Run(args []string) {
	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)
	socketPath := statusCmd.String("socket", ipc.DefaultSocketPath(), "Path to Unix socket")
	jsonOutput := statusCmd.Bool("json", false, "Output as JSON")

	if err := statusCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	client := ipc.NewClient(*socketPath)
	result, err := client.Status()
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
		fmt.Println("Server Status:")
		for k, v := range result {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}
}
