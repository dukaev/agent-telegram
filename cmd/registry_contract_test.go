package cmd

import (
	"strings"
	"testing"

	"agent-telegram/cmd/auth"
	"agent-telegram/cmd/chat"
	"agent-telegram/cmd/get"
	"agent-telegram/cmd/sys"
	"agent-telegram/internal/cliutil"
	"agent-telegram/internal/observability"
	"agent-telegram/internal/operations"
	telegramipc "agent-telegram/internal/telegram/ipc"

	"github.com/spf13/cobra"
)

func TestCommandMethodsHaveSchemasAndHandlers(t *testing.T) {
	handlerMethods := stringSet(telegramipc.RegisteredMethods())

	for _, method := range cliutil.CommandMethodNames() {
		if !cliutil.HasResultSchema(method) {
			t.Errorf("command method %q has no result schema", method)
		}
		if _, ok := handlerMethods[method]; !ok {
			t.Errorf("command method %q has no IPC handler", method)
		}
	}
}

func TestTelegramHandlersHaveResultSchemas(t *testing.T) {
	for _, method := range telegramipc.RegisteredMethods() {
		if !cliutil.HasResultSchema(method) {
			t.Errorf("IPC method %q has no result schema", method)
		}
	}
}

func TestOperationMethodsHaveHandlers(t *testing.T) {
	handlerMethods := stringSet(telegramipc.RegisteredMethods())
	for _, method := range []string{"ping", "echo", "status", "shutdown", "logout", "reload_session"} {
		handlerMethods[method] = struct{}{}
	}

	for _, method := range operations.Methods() {
		if _, ok := handlerMethods[method]; !ok {
			t.Errorf("operation method %q has no IPC handler", method)
		}
	}
}

func TestAgenticContractDoesNotExposeRemovedFlags(t *testing.T) {
	for _, name := range []string{"agent", "run-id"} {
		if RootCmd.PersistentFlags().Lookup(name) == nil {
			t.Fatalf("root should expose --%s for agent mode", name)
		}
	}
	for _, name := range []string{"text", "fields"} {
		if RootCmd.PersistentFlags().Lookup(name) != nil {
			t.Fatalf("root should not expose removed flag --%s", name)
		}
	}
	if auth.AuthCmd.Flags().Lookup("phone") != nil {
		t.Fatal("auth should not expose --phone")
	}
	for _, name := range []string{"web", "begin", "verify", "password", "status"} {
		if childCommand(auth.AuthCmd, name) != nil {
			t.Fatalf("auth should not expose legacy subcommand %q", name)
		}
	}
	if observability.ParseRedactionMode("full") != observability.RedactionSafe {
		t.Fatal("redaction=full should not be exposed as an effective mode")
	}
	if sys.AuditCmd.Flags().Lookup("redaction").DefValue != "safe" {
		t.Fatal("audit redaction should default to safe")
	}
	if childCommand(RootCmd, "login") != nil {
		t.Fatal("root should not expose removed login alias")
	}
	for name, cmd := range map[string]*cobra.Command{
		"status":    sys.StatusCmd,
		"updates":   get.UpdatesCmd,
		"chat open": chat.OpenCmd,
	} {
		if cmd.Flags().Lookup("json") != nil {
			t.Fatalf("%s should not expose removed --json flag", name)
		}
	}
}

func TestFirstArgPeerCommandRegistry(t *testing.T) {
	accepted := []string{
		"bot step",
		"bot press",
		"send",
		"send text",
		"send poll",
		"send contact",
		"send location",
		"send dice",
		"game dice",
		"chat keyboard",
		"msg list",
		"msg replies",
		"msg inspect-keyboard",
		"msg press-keyboard",
		"msg wait",
		"msg reply-comment",
	}
	for _, path := range accepted {
		command := commandAtPath(RootCmd, path)
		if command == nil {
			t.Errorf("command %q is not registered", path)
			continue
		}
		if !cliutil.AcceptsFirstArgPeer(command) {
			t.Errorf("command %q should accept a first positional peer", path)
		}
	}

	for _, path := range []string{"msg get", "msg press-button", "send update"} {
		command := commandAtPath(RootCmd, path)
		if command == nil {
			t.Errorf("command %q is not registered", path)
			continue
		}
		if cliutil.AcceptsFirstArgPeer(command) {
			t.Errorf("command %q must not treat its first positional number as a peer", path)
		}
	}
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

func childCommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, child := range cmd.Commands() {
		if child.Name() == name {
			return child
		}
	}
	return nil
}

func commandAtPath(root *cobra.Command, path string) *cobra.Command {
	command := root
	for _, name := range strings.Fields(path) {
		command = childCommand(command, name)
		if command == nil {
			return nil
		}
	}
	return command
}
