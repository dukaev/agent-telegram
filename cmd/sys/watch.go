// Package sys provides system commands.
package sys

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// WatchCmd represents the watch command.
var WatchCmd = &cobra.Command{
	GroupID: "server",
	Use:     "watch",
	Short:   "Watch server logs in real-time",
	Long: `Watch the server logs in real-time (like docker logs -f).
Follows the agent-telegram.log file and outputs new lines as they are added.

Example:
  agent-telegram watch`,
	Run: func(_ *cobra.Command, _ []string) {
		logFile := "agent-telegram.log"

		// Open log file
		f, err := os.Open(logFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot open log file: %v\n", err)
			fmt.Fprintln(os.Stderr, "Make sure the server is running in daemon mode")
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()

		// Seek to end of file
		if _, err := f.Seek(0, 2); err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot seek in log file: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Watching logs (Ctrl+C to stop)...")

		// Read file line by line using polling
		reader := bufio.NewReader(f)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			line, err := reader.ReadString('\n')
			if err != nil {
				continue
			}
			fmt.Print(line)
		}
	},
}

// AddWatchCommand adds the watch command to the root command.
func AddWatchCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(WatchCmd)
}
