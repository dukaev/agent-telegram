// Package chat provides commands for managing chats.
package chat

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	createChannelTitle       string
	createChannelDescription string
	createChannelUsername    string
	createChannelMegagroup   bool
)

// CreateChannelCmd represents the create-channel command.
var CreateChannelCmd = &cobra.Command{
	Use:     "create-channel",
	Short:   "Create a new channel or supergroup",
	Long: `Create a new Telegram channel or supergroup with the specified title.

Use --title to set the channel title.
Use --description to set the channel description (optional).
Use --username to set a public username (optional).
Use --megagroup to create a supergroup instead of a channel.

Example:
  agent-telegram create-channel --title "My Channel" --description "My channel description"`,
	Args: cobra.NoArgs,
}

// AddCreateChannelCommand adds the create-channel command to the root command.
func AddCreateChannelCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(CreateChannelCmd)

	CreateChannelCmd.Flags().StringVarP(&createChannelTitle, "title", "t", "", "Channel title")
	CreateChannelCmd.Flags().StringVarP(&createChannelDescription, "description", "d", "", "Channel description")
	CreateChannelCmd.Flags().StringVarP(&createChannelUsername, "username", "u", "", "Public username")
	CreateChannelCmd.Flags().BoolVarP(&createChannelMegagroup, "megagroup", "g", false, "Create as supergroup")
	_ = CreateChannelCmd.MarkFlagRequired("title")

	CreateChannelCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(CreateChannelCmd, true)
		params := map[string]any{
			"title":       createChannelTitle,
			"description": createChannelDescription,
			"username":    createChannelUsername,
			"megagroup":   createChannelMegagroup,
		}

		result := runner.CallWithParams("create_channel", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)

		// Print human-readable summary
		cliutil.PrintSuccessSummary(result, "Channel created successfully")
		cliutil.PrintResultField(result, "title", "  Title: %s\n")
		cliutil.PrintResultField(result, "chatId", "  Chat ID: %d\n")
	}
}
