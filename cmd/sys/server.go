package sys

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
	"agent-telegram/internal/ipc"
)

var (
	serverWaitTimeout time.Duration
	serverPoll        time.Duration
)

// ServerCmd groups explicit daemon lifecycle helpers for agents.
var ServerCmd = &cobra.Command{
	GroupID: "server",
	Use:     "server",
	Short:   "Manage the IPC server lifecycle",
}

// ServerEnsureCmd starts the daemon if needed and waits until it is ready.
var ServerEnsureCmd = &cobra.Command{
	Use:   "ensure",
	Short: "Start IPC server if needed and wait until ready",
	Run:   runServerEnsure,
}

// ServerWaitReadyCmd waits for an already-starting server to become ready.
var ServerWaitReadyCmd = &cobra.Command{
	Use:   "wait-ready",
	Short: "Wait until IPC server is ready",
	Run:   runServerWaitReady,
}

// AddServerCommand adds explicit server lifecycle commands.
func AddServerCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ServerCmd)
	ServerCmd.AddCommand(ServerEnsureCmd, ServerWaitReadyCmd)

	for _, cmd := range []*cobra.Command{ServerEnsureCmd, ServerWaitReadyCmd} {
		cmd.Flags().DurationVar(&serverWaitTimeout, "timeout", 30*time.Second, "Maximum time to wait")
		cmd.Flags().DurationVar(&serverPoll, "interval", 250*time.Millisecond, "Polling interval")
	}
}

func runServerEnsure(cmd *cobra.Command, _ []string) {
	runner := cliutil.NewRunnerFromCmd(cmd, true)
	socketPath, _ := cmd.Flags().GetString("socket")
	client := ipc.NewClient(socketPath)
	if status, err := client.Call("status", nil); err == nil && serverReady(status) {
		runner.PrintJSON(map[string]any{"ok": true, "started": false, "status": status})
		return
	}

	exe, err := os.Executable()
	if err != nil {
		runner.PrintJSON(map[string]any{"ok": false, "error": err.Error()})
		cliutil.Exit(1)
	}
	args := []string{"serve", "--foreground", "--socket", socketPath}
	//nolint:gosec,noctx // exe is the current binary and serve is intentionally detached.
	start := exec.Command(exe, args...)
	start.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	devNull, devNullErr := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if devNullErr == nil {
		defer func() { _ = devNull.Close() }()
		start.Stdin = devNull
		start.Stdout = devNull
		start.Stderr = devNull
	}
	if err := start.Start(); err != nil {
		runner.PrintJSON(map[string]any{"ok": false, "error": err.Error()})
		cliutil.Exit(1)
	}

	status, err := waitReady(client, serverWaitTimeout, serverPoll)
	if err != nil {
		runner.PrintJSON(map[string]any{"ok": false, "started": true, "pid": start.Process.Pid, "error": err.Error()})
		cliutil.Exit(1)
	}
	runner.PrintJSON(map[string]any{"ok": true, "started": true, "pid": start.Process.Pid, "status": status})
}

func runServerWaitReady(cmd *cobra.Command, _ []string) {
	runner := cliutil.NewRunnerFromCmd(cmd, true)
	socketPath, _ := cmd.Flags().GetString("socket")
	status, err := waitReady(ipc.NewClient(socketPath), serverWaitTimeout, serverPoll)
	if err != nil {
		runner.PrintJSON(map[string]any{"ok": false, "error": err.Error()})
		cliutil.Exit(1)
	}
	runner.PrintJSON(map[string]any{"ok": true, "status": status})
}

func waitReady(client *ipc.Client, timeout, interval time.Duration) (any, error) {
	if interval <= 0 {
		interval = 250 * time.Millisecond
	}
	deadline := time.Now().Add(timeout)
	var lastErr string
	readyCount := 0
	var readyStatus any
	for time.Now().Before(deadline) {
		status, err := client.Call("status", nil)
		if err == nil && serverReady(status) {
			readyCount++
			readyStatus = status
			if readyCount >= 2 {
				return readyStatus, nil
			}
			time.Sleep(interval)
			continue
		}
		readyCount = 0
		if err != nil {
			lastErr = err.Message
		} else if data, jsonErr := json.Marshal(status); jsonErr == nil {
			lastErr = string(data)
		} else {
			lastErr = fmt.Sprint(status)
		}
		time.Sleep(interval)
	}
	return nil, fmt.Errorf("server not ready within %s: %s", timeout, lastErr)
}

func serverReady(status any) bool {
	m, ok := status.(map[string]any)
	if !ok {
		return false
	}
	initialized, _ := m["initialized"].(bool)
	return initialized
}
