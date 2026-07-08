package folders

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAddFoldersCommandRegistersExpectedSurface(t *testing.T) {
	root := &cobra.Command{Use: "root"}

	AddFoldersCommand(root)

	foldersCmd := childCommand(root, "folders")
	if foldersCmd == nil {
		t.Fatal("folders command was not registered")
	}
	for _, name := range []string{"list", "create", "delete"} {
		if childCommand(foldersCmd, name) == nil {
			t.Fatalf("folders subcommand %q was not registered", name)
		}
	}
	for _, flag := range []string{"title", "include", "exclude", "include-groups", "include-channels", "include-contacts"} {
		if CreateCmd.Flags().Lookup(flag) == nil {
			t.Fatalf("create flag --%s was not registered", flag)
		}
	}
	if DeleteCmd.Flags().Lookup("id") == nil {
		t.Fatal("delete flag --id was not registered")
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
