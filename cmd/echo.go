// Package cmd provides CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

// echoCmd represents the echo command.
var echoCmd = &cobra.Command{
	Use:   "echo <message>",
	Short: "Echo message via IPC server",
	Long:  `Send a message to the IPC server and have it echoed back.`,
	Args:  cobra.ExactArgs(1),
	Run:   runEcho,
}

func init() {
	rootCmd.AddCommand(echoCmd)
}

func runEcho(cmd *cobra.Command, args []string) {
	socketPath, _ := cmd.Flags().GetString("socket")
	message := args[0]

	client := ipc.NewClient(socketPath)
	result, err := client.Echo(message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result)
}
