// Package worker provides the worker command implementation.
package worker

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"agent-telegram/internal/ipc"
)

// Run executes the worker command with the given arguments.
func Run(args []string) {
	workerCmd := flag.NewFlagSet("worker", flag.ExitOnError)
	sessionPath := workerCmd.String("session", "", "Path to session file (default: auto-detect)")

	if err := workerCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down worker...")
		cancel()
	}()

	fmt.Println("Starting worker...")
	_ = sessionPath

	// Create and configure IPC server
	srv := ipc.NewServer()
	ipc.RegisterPingPong(srv)

	// Start IPC server in a goroutine
	go func() {
		if err := srv.ServeStdinStdout(); err != nil {
			log.Printf("IPC server error: %v", err)
			cancel()
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	fmt.Println("\nWorker stopped.")
	cancel()
}
