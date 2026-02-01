// Package sys provides system commands.
package sys

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var statusJSON bool

// StatusCmd represents the status command.
var StatusCmd = &cobra.Command{
	GroupID: "server",
	Use:     "status",
	Short:   "Check if the IPC server is running",
	Long: `Check the status of the IPC server.
Returns server status and PID if running, or error if not reachable.

Example:
  agent-telegram status`,
}

// AddStatusCommand adds the status command to the root command.
func AddStatusCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(StatusCmd)

	StatusCmd.Flags().BoolVarP(&statusJSON, "json", "j", false, "Output as JSON")

	StatusCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(StatusCmd, true)

		// Use Client directly to avoid auto-starting the server
		client := runner.Client()
		result, err := client.Call("status", nil)
		if err != nil {
			fmt.Println("Server: not running")
			return
		}

		m, ok := cliutil.ToMap(result)
		if !ok {
			fmt.Println("Server: unknown state")
			return
		}

		if statusJSON {
			runner.PrintJSON(result)
			return
		}

		// Server status
		status := cliutil.ExtractString(m, "status")
		if status == "running" {
			if pid, ok := m["pid"].(float64); ok {
				fmt.Printf("Server: running (PID %d)\n", int64(pid))
			} else {
				fmt.Println("Server: running")
			}
		} else {
			fmt.Println("Server: not running")
			return
		}

		// Session path
		if sessionPath := cliutil.ExtractString(m, "session_path"); sessionPath != "" {
			fmt.Printf("Session: %s\n", sessionPath)
		}

		// Telegram status
		initialized, _ := m["initialized"].(bool)
		authorized, _ := m["authorized"].(bool)

		//nolint:gocritic // ifElseChain is clearer than switch for boolean conditions
		if !initialized {
			fmt.Println("Telegram: initializing...")
		} else if !authorized {
			fmt.Println("Telegram: not authorized (run: agent-telegram login)")
		} else {
			username := cliutil.ExtractString(m, "username")
			firstName := cliutil.ExtractString(m, "firstName")
			switch {
			case username != "":
				fmt.Printf("Telegram: authorized (@%s)\n", username)
			case firstName != "":
				fmt.Printf("Telegram: authorized (%s)\n", firstName)
			default:
				fmt.Println("Telegram: authorized")
			}
		}
	}
}
