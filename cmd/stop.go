// Package cmd provides CLI commands.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

var (
	stopSocket string
	stopForce  bool
)

// stopCmd represents the stop command.
var stopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stop the IPC server",
	Long:    `Stop the background IPC server gracefully.`,
	Run:     runStop,
	GroupID: GroupIDServer,
}

func init() {
	RootCmd.AddCommand(stopCmd)

	stopCmd.Flags().StringVarP(&stopSocket, "socket", "s", "",
		"Path to Unix socket (default: /tmp/agent-telegram.sock)")
	stopCmd.Flags().BoolVarP(&stopForce, "force", "f", false,
		"Force stop by sending SIGKILL instead of SIGTERM")
}

func runStop(_ *cobra.Command, _ []string) {
	socketPath := getStopSocketPath()

	client := ipc.NewClient(socketPath)

	// Try to get PID first
	var pid int
	result, err := client.Call("status", nil)
	if err == nil {
		if r, ok := result.(map[string]any); ok {
			if p, ok := r["pid"].(float64); ok {
				pid = int(p)
			}
		}
	}

	if stopForce {
		if pid > 0 {
			process, err := os.FindProcess(pid)
			if err == nil {
				_ = process.Kill()
				fmt.Printf("Force killed server (PID %d)\n", pid)
				return
			}
		}
		fmt.Fprintln(os.Stderr, "Could not find process to force kill")
		os.Exit(1)
	}

	// Send shutdown command
	result, err = client.Call("shutdown", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop server: %v\n", err)
		fmt.Fprintln(os.Stderr, "Make sure the server is running")
		os.Exit(1)
	}

	// Parse response
	var resp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if resultBytes, err := json.Marshal(result); err == nil {
		_ = json.Unmarshal(resultBytes, &resp)
	}

	if resp.Message != "" {
		fmt.Println(resp.Message)
	} else {
		fmt.Println("Server stopped successfully")
	}

	// Give it a moment to shut down gracefully
	time.Sleep(500 * time.Millisecond)
}

// getStopSocketPath returns the socket path from flags or default.
func getStopSocketPath() string {
	socketPath, _ := RootCmd.Flags().GetString("socket")
	if stopSocket != "" {
		socketPath = stopSocket
	}
	if socketPath == "" {
		socketPath = "/tmp/agent-telegram.sock"
	}
	return socketPath
}
