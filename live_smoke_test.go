package main

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestLiveSmokeReadOnly(t *testing.T) {
	if os.Getenv("AGENT_TELEGRAM_LIVE_TESTS") != "1" {
		t.Skip("set AGENT_TELEGRAM_LIVE_TESTS=1 to run read-only live smoke tests")
	}

	for _, args := range [][]string{
		{"run", ".", "auth", "status"},
		{"run", ".", "manifest"},
		{"run", ".", "my-info", "--summary"},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		cmd := exec.CommandContext(ctx, "go", args...)
		out, err := cmd.CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("go %v failed: %v\n%s", args, err, out)
		}
	}
}
