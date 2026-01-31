// Package ping provides the ping command implementation.
package ping

import (
	"flag"
	"fmt"
	"os"

	"agent-telegram/internal/ipc"
)

// Run executes the ping command.
func Run(args []string) {
	pingCmd := flag.NewFlagSet("ping", flag.ExitOnError)
	socketPath := pingCmd.String("socket", ipc.DefaultSocketPath(), "Path to Unix socket")
	message := pingCmd.String("message", "hello", "Message to send")

	if err := pingCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	client := ipc.NewClient(*socketPath)
	result, err := client.Ping(*message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Pong: %s (pong: %t)\n", result.Message, result.Pong)
}
