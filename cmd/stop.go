// Package cmd provides CLI commands.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
	"agent-telegram/internal/paths"
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

	// Try to get PID from file first, then from RPC
	pid := getPIDFromFile()
	if pid == 0 {
		pid = getPIDFromRPC(client)
	}

	if stopForce {
		if pid > 0 {
			process, err := os.FindProcess(pid)
			if err == nil {
				_ = process.Kill()
				fmt.Printf("Force killed server (PID %d)\n", pid)
				cleanupPIDFile()
				return
			}
		}
		fmt.Fprintln(os.Stderr, "Could not find process to force kill")
		os.Exit(1)
	}

	// Send shutdown command
	result, err := client.Call("shutdown", nil)
	if err != nil {
		// If RPC fails but we have PID, suggest force kill
		if pid > 0 {
			fmt.Fprintf(os.Stderr, "Failed to stop server gracefully: %v\n", err)
			fmt.Fprintf(os.Stderr, "Try: agent-telegram stop --force (PID %d)\n", pid)
		} else {
			fmt.Fprintf(os.Stderr, "Failed to stop server: %v\n", err)
			fmt.Fprintln(os.Stderr, "Make sure the server is running")
		}
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

	// Wait for server to shut down
	waitForShutdown(client, 5*time.Second)
}

// getPIDFromFile reads PID from the PID file.
func getPIDFromFile() int {
	pidPath, err := paths.PIDFilePath()
	if err != nil {
		return 0
	}
	pid, err := paths.ReadPID(pidPath)
	if err != nil {
		return 0
	}
	return pid
}

// getPIDFromRPC gets PID from the running server via RPC.
func getPIDFromRPC(client *ipc.Client) int {
	result, err := client.Call("status", nil)
	if err != nil {
		return 0
	}
	if r, ok := result.(map[string]any); ok {
		if p, ok := r["pid"].(float64); ok {
			return int(p)
		}
	}
	return 0
}

// cleanupPIDFile removes the PID file after force kill.
func cleanupPIDFile() {
	pidPath, err := paths.PIDFilePath()
	if err != nil {
		return
	}
	_ = paths.RemovePID(pidPath)
}

// waitForShutdown waits for the server to shut down gracefully.
func waitForShutdown(client *ipc.Client, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, err := client.Call("status", nil)
		if err != nil {
			// Server is down
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
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
