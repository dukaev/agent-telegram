// Package cmd provides CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

var (
	pingMessage string
)

// pingCmd represents the ping command.
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Send ping to IPC server",
	Long:  `Send a ping message to the IPC server and receive a pong response.`,
	Run:   runPing,
}

func init() {
	rootCmd.AddCommand(pingCmd)

	pingCmd.Flags().StringVarP(&pingMessage, "message", "m", "hello", "Message to send")
}

func runPing(_ *cobra.Command, _ []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")

	client := ipc.NewClient(socketPath)
	result, err := client.Ping(pingMessage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Pong: %s (pong: %t)\n", result.Message, result.Pong)
}
