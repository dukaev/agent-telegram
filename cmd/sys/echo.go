// Package sys provides system commands.
package sys

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

// EchoCmd represents the echo command.
var EchoCmd = &cobra.Command{
	Use:   "echo <message>",
	Short: "Echo message via IPC server",
	Long:  `Send a message to the IPC server and have it echoed back.`,
	Args:  cobra.ExactArgs(1),
}

// AddEchoCommand adds the echo command to the root command.
func AddEchoCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(EchoCmd)

	EchoCmd.Run = func(command *cobra.Command, args []string) {
		socketPath, _ := command.Flags().GetString("socket")
		message := args[0]

		client := ipc.NewClient(socketPath)
		result, err := client.Echo(message)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(result)
	}
}
