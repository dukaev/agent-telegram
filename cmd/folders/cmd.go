// Package folders provides commands for managing chat folders.
package folders

import (
	"github.com/spf13/cobra"
)

// FoldersCmd represents the folders command group.
var FoldersCmd = &cobra.Command{
	Use:   "folders",
	Short: "Manage chat folders",
	Long:  `Commands for managing Telegram chat folders (filters).`,
}

// AddFoldersCommand adds the folders command group to the root command.
func AddFoldersCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(FoldersCmd)

	AddListCommand(FoldersCmd)
	AddCreateCommand(FoldersCmd)
	AddDeleteCommand(FoldersCmd)
}
