// Package main provides a simple IPC client for testing.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	// Start worker process
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "./agent-telegram", "worker")

	// Create pipes for stdin/stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to create stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to create stdout pipe: %v", err)
	}

	// Start the worker
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}
	defer func() { _ = cmd.Wait() }()

	fmt.Println("Worker started. Enter JSON-RPC requests (one per line):")
	fmt.Println(`Example: {"jsonrpc":"2.0","method":"ping","params":{"message":"hello"},"id":1}`)
	fmt.Println("Press Ctrl+C to exit")

	// Read responses in a goroutine
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Printf("← Response: %s\n", scanner.Text())
		}
	}()

	// Send requests from stdin
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(stdin)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req map[string]interface{}
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			fmt.Printf("✗ Invalid JSON: %v\n", err)
			continue
		}

		fmt.Printf("→ Request: %s\n", line)
		if err := encoder.Encode(req); err != nil {
			fmt.Printf("✗ Failed to send request: %v\n", err)
		}
	}
}
