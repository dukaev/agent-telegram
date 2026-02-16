// Package game provides interactive game commands.
package game

import "github.com/spf13/cobra"

// GameCmd represents the parent game command.
var GameCmd = &cobra.Command{
	GroupID: "chat",
	Use:     "game",
	Short:   "Play games in Telegram chats",
}

// AddGameCommand adds the game command and all subcommands to the root command.
func AddGameCommand(rootCmd *cobra.Command) {
	addDiceCommand(GameCmd)
	rootCmd.AddCommand(GameCmd)
}
