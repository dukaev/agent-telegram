// Package cmd provides CLI commands.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

var (
	statusJSON bool
)

// statusCmd represents the status command.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check IPC server status",
	Long:  `Check if the IPC server is running and get its status information.`,
	Run: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().BoolVarP(&statusJSON, "json", "j", false, "Output as JSON")
}

func runStatus(_ *cobra.Command, _ []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")

	client := ipc.NewClient(socketPath)
	result, err := client.Status()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if statusJSON {
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
