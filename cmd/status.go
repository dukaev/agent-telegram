// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	statusJSON bool
)

// statusCmd represents the status command.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check IPC server status",
	Long:  `Check if the IPC server is running and get its status information.`,
	Run:   runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().BoolVarP(&statusJSON, "json", "j", false, "Output as JSON")
}

func runStatus(_ *cobra.Command, _ []string) {
	runner := NewRunnerFromRoot(statusJSON)

	result := runner.Call("status", nil)

	if statusJSON {
		runner.PrintJSON(result)
	} else {
		printStatus(result)
	}
}

// printStatus prints status in human-readable format.
func printStatus(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		return
	}

	fmt.Println("Server Status:")
	for k, v := range r {
		fmt.Printf("  %s: %v\n", k, v)
	}
}
