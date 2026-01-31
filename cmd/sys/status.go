// Package sys provides system commands.
package sys

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	statusJSON bool
)

// StatusCmd represents the status command.
var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check IPC server status",
	Long:  `Check if the IPC server is running and get its status information.`,
}

// AddStatusCommand adds the status command to the root command.
func AddStatusCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(StatusCmd)

	StatusCmd.Flags().BoolVarP(&statusJSON, "json", "j", false, "Output as JSON")

	StatusCmd.Run = func(*cobra.Command, []string) {
		runner := cliutil.NewRunnerFromCmd(StatusCmd, statusJSON)

		result := runner.Call("status", nil)

		if statusJSON {
			runner.PrintJSON(result)
		} else {
			printStatus(result)
		}
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
