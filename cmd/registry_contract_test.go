package cmd

import (
	"testing"

	"agent-telegram/cmd/auth"
	"agent-telegram/cmd/sys"
	"agent-telegram/internal/cliutil"
	"agent-telegram/internal/observability"
	"agent-telegram/internal/operations"
	telegramipc "agent-telegram/internal/telegram/ipc"
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

	for _, method := range operations.Methods() {
		if _, ok := handlerMethods[method]; !ok {
			t.Errorf("operation method %q has no IPC handler", method)
		}
	}
}

func TestAgenticContractDoesNotExposeRemovedFlags(t *testing.T) {
	for _, name := range []string{"text", "fields"} {
		if RootCmd.PersistentFlags().Lookup(name) != nil {
			t.Fatalf("root should not expose removed flag --%s", name)
		}
	}
	if auth.AuthWebCmd.Flags().Lookup("phone") != nil {
		t.Fatal("auth web should not expose --phone")
	}
	if auth.AuthBeginCmd.Flags().Lookup("phone") != nil {
		t.Fatal("auth begin should not expose --phone")
	}
	if observability.ParseRedactionMode("full") != observability.RedactionSafe {
		t.Fatal("redaction=full should not be exposed as an effective mode")
	}
	if sys.AuditCmd.Flags().Lookup("redaction").DefValue != "safe" {
		t.Fatal("audit redaction should default to safe")
	}
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}
