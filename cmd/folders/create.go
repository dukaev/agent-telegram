// Package folders provides commands for managing chat folders.
package folders

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	createTitle           string
	createIncludedChats   []string
	createExcludedChats   []string
	createIncludeContacts bool
	createIncludeGroups   bool
	createIncludeChannels bool
	createIncludeBots     bool
)

// CreateCmd represents the folders create command.
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new chat folder",
	Long: `Create a new chat folder with specified settings.

Example:
  agent-telegram folders create --title "Work" --include-groups --include-channels
  agent-telegram folders create --title "Friends" --include-contacts
  agent-telegram folders create --title "Custom" --include @user1 --include @channel1`,
	Args: cobra.NoArgs,
}

// AddCreateCommand adds the create command to the parent command.
func AddCreateCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(CreateCmd)

	CreateCmd.Flags().StringVarP(&createTitle, "title", "T", "", "Folder title")
	CreateCmd.Flags().StringSliceVar(&createIncludedChats, "include", nil, "Chats to include")
	CreateCmd.Flags().StringSliceVar(&createExcludedChats, "exclude", nil, "Chats to exclude")
	CreateCmd.Flags().BoolVar(&createIncludeContacts, "include-contacts", false, "Include contacts")
	CreateCmd.Flags().BoolVar(&createIncludeGroups, "include-groups", false, "Include groups")
	CreateCmd.Flags().BoolVar(&createIncludeChannels, "include-channels", false, "Include channels")
	CreateCmd.Flags().BoolVar(&createIncludeBots, "include-bots", false, "Include bots")
	_ = CreateCmd.MarkFlagRequired("title")

	CreateCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(CreateCmd, true)
		params := map[string]any{
			"title":           createTitle,
			"includeContacts": createIncludeContacts,
			"includeGroups":   createIncludeGroups,
			"includeChannels": createIncludeChannels,
			"includeBots":     createIncludeBots,
		}
		if len(createIncludedChats) > 0 {
			params["includedChats"] = createIncludedChats
		}
		if len(createExcludedChats) > 0 {
			params["excludedChats"] = createExcludedChats
		}

		result := runner.CallWithParams("create_folder", params)
		runner.PrintResult(result, func(any) {
			cliutil.PrintSuccessSummary(result, "Folder created")
		})
	}
}
