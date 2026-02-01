// Package folders provides commands for managing chat folders.
package folders

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// DeleteCmd represents the folders delete command.
var DeleteCmd = cliutil.NewSimpleCommand(cliutil.SimpleCommandDef{
	Use:   "delete",
	Short: "Delete a chat folder",
	Long: `Delete a chat folder by its ID.

Use 'folders list' to see folder IDs.

Example:
  agent-telegram folders delete --id 2`,
	Method: "delete_folder",
	Flags: []cliutil.Flag{
		{Name: "id", Short: "i", Usage: "Folder ID to delete", Required: true, Type: cliutil.FlagInt},
	},
	Success: "Folder deleted",
})

// AddDeleteCommand adds the delete command to the parent command.
func AddDeleteCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(DeleteCmd)
}
