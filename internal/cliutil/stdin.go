package cliutil

import (
	"io"
	"os"
	"strings"
)

// ReadStdinIfPiped reads all of stdin if it is piped (not a TTY).
// Returns "" if stdin is a terminal (never blocks waiting for input).
func ReadStdinIfPiped() string {
	info, err := os.Stdin.Stat()
	if err != nil {
		return ""
	}
	// If stdin is a character device (TTY), don't read
	if info.Mode()&os.ModeCharDevice != 0 {
		return ""
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(data), "\n")
}
