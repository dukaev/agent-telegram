// Package chat provides commands for managing chats.
package chat

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	createGroupTitle   string
	createGroupMembers []string
)

// CreateGroupCmd represents the create-group command.
var CreateGroupCmd = &cobra.Command{
	Use:     "create-group",
	Short:   "Create a new group chat",
	Long: `Create a new Telegram group chat with the specified title and members.

Use --title to set the group title.
Use --members to specify users to add (can be specified multiple times).

Example:
  agent-telegram create-group --title "My Group" --members @user1 --members @user2`,
	Args: cobra.NoArgs,
}

// AddCreateGroupCommand adds the create-group command to the root command.
func AddCreateGroupCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(CreateGroupCmd)

	CreateGroupCmd.Flags().StringVarP(&createGroupTitle, "title", "t", "", "Group title")
	CreateGroupCmd.Flags().StringSliceVarP(&createGroupMembers, "members", "m", []string{},
		"Group members (can be specified multiple times)")
	_ = CreateGroupCmd.MarkFlagRequired("title")
	_ = CreateGroupCmd.MarkFlagRequired("members")

	CreateGroupCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(CreateGroupCmd, true)
		params := map[string]any{
			"title":   createGroupTitle,
			"members": createGroupMembers,
		}

		result := runner.CallWithParams("create_group", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)

		// Print human-readable summary
		cliutil.PrintSuccessSummary(result, "Group created successfully")
		cliutil.PrintResultField(result, "title", "  Title: %s\n")
		cliutil.PrintResultField(result, "chatId", "  Chat ID: %d\n")
	}
}
