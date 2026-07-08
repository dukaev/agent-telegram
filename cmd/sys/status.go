// Package sys provides system commands.
package sys

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

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

	StatusCmd.Run = func(cmd *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(cmd, true)
		runner.SetAction("status")
		jsonStatus := statusWantsJSON(cmd, runner)

		// Use Client directly so status can report a stopped server.
		client := runner.Client()
		result, err := client.Call("status", nil)
		if err != nil {
			if jsonStatus {
				runner.PrintJSON(map[string]any{
					"ok":     false,
					"status": "not_running",
					"error":  err.Message,
				})
				return
			}
			fmt.Println("Server: not running")
			return
		}

		m, ok := cliutil.ToMap(result)
		if !ok {
			if jsonStatus {
				runner.PrintJSON(map[string]any{
					"ok":     false,
					"status": "unknown",
				})
				return
			}
			fmt.Println("Server: unknown state")
			return
		}

		if jsonStatus {
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

		// Session storage
		if sessionStorage := cliutil.ExtractString(m, "session_storage"); sessionStorage != "" {
			fmt.Printf("Session: %s\n", sessionStorage)
		}

		// Telegram status
		initialized, _ := m["initialized"].(bool)
		authorized, _ := m["authorized"].(bool)
		state := cliutil.ExtractString(m, "telegram_state")

		//nolint:gocritic // ifElseChain is clearer than switch for boolean conditions
		if !initialized {
			if state == "" {
				state = "initializing"
			}
			fmt.Printf("Telegram: %s...\n", state)
		} else if !authorized {
			fmt.Println("Telegram: not authorized (run: agent-telegram auth web)")
		} else {
			username := cliutil.ExtractString(m, "username")
			firstName := cliutil.ExtractString(m, "first_name")
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

func statusWantsJSON(cmd *cobra.Command, runner *cliutil.Runner) bool {
	if runner.AgentMode() {
		return true
	}
	output, _ := cmd.Flags().GetString("output")
	return strings.EqualFold(output, "json")
}
