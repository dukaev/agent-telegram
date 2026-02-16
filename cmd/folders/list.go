// Package folders provides commands for managing chat folders.
package folders

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// ListCmd represents the folders list command.
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all chat folders",
	Long: `List all chat folders (filters) in your Telegram account.

Example:
  agent-telegram folders list`,
	Args: cobra.NoArgs,
}

// AddListCommand adds the list command to the parent command.
func AddListCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ListCmd)

	ListCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(ListCmd, true)
		runner.SetIDKey("id")

		result := runner.Call("get_folders", nil)
		runner.PrintResult(result, func(r any) {
			if m, ok := r.(map[string]any); ok {
				if folders, ok := m["folders"].([]any); ok {
					fmt.Fprintf(os.Stderr, "Found %d folder(s)\n", len(folders))
					for _, f := range folders {
						if folder, ok := f.(map[string]any); ok {
							id := cliutil.ExtractFloat64(folder, "id")
							title := cliutil.ExtractString(folder, "title")
							fmt.Fprintf(os.Stderr, "  [%d] %s\n", int(id), title)
						}
					}
				}
			}
		})
	}
}
