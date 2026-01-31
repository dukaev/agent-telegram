// Package echo provides the echo command implementation.
package echo

import (
	"flag"
	"fmt"
	"os"

	"agent-telegram/internal/ipc"
)

// Run executes the echo command.
func Run(args []string) {
	echoCmd := flag.NewFlagSet("echo", flag.ExitOnError)
	socketPath := echoCmd.String("socket", ipc.DefaultSocketPath(), "Path to Unix socket")

	if err := echoCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Get message from remaining args
	if echoCmd.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: agent-telegram echo <message>")
		os.Exit(1)
	}
	message := echoCmd.Arg(0)

	client := ipc.NewClient(*socketPath)
	result, err := client.Echo(message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result)
}
