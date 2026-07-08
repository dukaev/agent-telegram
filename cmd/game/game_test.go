package game

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAddGameCommandRegistersDice(t *testing.T) {
	root := &cobra.Command{Use: "root"}

	AddGameCommand(root)

	gameCmd := childCommand(root, "game")
	if gameCmd == nil {
		t.Fatal("game command was not registered")
	}
	diceCmd := childCommand(gameCmd, "dice")
	if diceCmd == nil {
		t.Fatal("dice command was not registered")
	}
	if diceCmd.Flags().Lookup("to") == nil || diceCmd.Flags().Lookup("emoticon") == nil {
		t.Fatal("expected dice flags")
	}
}

func childCommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, child := range cmd.Commands() {
		if child.Name() == name {
			return child
		}
	}
	return nil
}
