// Package sys provides system commands.
package sys

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

var (
	pingMessage string
)

// PingCmd represents the ping command.
var PingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Send ping to IPC server",
	Long:  `Send a ping message to the IPC server and receive a pong response.`,
}

// AddPingCommand adds the ping command to the root command.
func AddPingCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(PingCmd)

	PingCmd.Flags().StringVarP(&pingMessage, "message", "m", "hello", "Message to send")

	PingCmd.Run = func(command *cobra.Command, _ []string) {
		socketPath, _ := command.Flags().GetString("socket")

		client := ipc.NewClient(socketPath)
		result, err := client.Ping(pingMessage)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Pong: %s (pong: %t)\n", result.Message, result.Pong)
	}
}
