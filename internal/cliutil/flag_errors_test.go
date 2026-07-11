package cliutil

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestFlagErrorWithHintsSuggestsNearestFlagAndExample(t *testing.T) {
	root := &cobra.Command{Use: "agent-telegram"}
	bot := &cobra.Command{Use: "bot"}
	step := &cobra.Command{
		Use:     "step <peer>",
		Example: "  agent-telegram bot step <peer> --send <text>",
	}
	step.Flags().String("send", "", "")
	step.Flags().String("text", "", "")
	root.AddCommand(bot)
	bot.AddCommand(step)
	root.SetFlagErrorFunc(FlagErrorWithHints)
	root.SetArgs([]string{"bot", "step", "@bot", "--txt", "hello"})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil")
	}
	message := err.Error()
	if !strings.Contains(message, "unknown flag: --txt") {
		t.Fatalf("error = %q, want original unknown flag", message)
	}
	if !strings.Contains(message, "did you mean --text?") {
		t.Fatalf("error = %q, want --text suggestion", message)
	}
	if !strings.Contains(message, "agent-telegram bot step <peer> --send <text>") {
		t.Fatalf("error = %q, want canonical example", message)
	}
}

func TestFlagErrorWithHintsExplainsNegativePeerID(t *testing.T) {
	cmd := &cobra.Command{Use: "get"}
	err := FlagErrorWithHints(cmd, errors.New("unknown shorthand flag: '5' in -5424738551"))
	if err == nil || !strings.Contains(err.Error(), "negative peer IDs must use --to=-5424738551") {
		t.Fatalf("error = %v, want negative peer guidance", err)
	}
}

func TestFlagErrorWithHintsLeavesUnrelatedErrorsUnchanged(t *testing.T) {
	want := errors.New("required flag missing")
	if got := FlagErrorWithHints(&cobra.Command{Use: "test"}, want); got != want {
		t.Fatalf("error = %v, want original %v", got, want)
	}
}

func TestFlagErrorWithHintsRequiresUniqueNearestFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("cat", "", "")
	cmd.Flags().String("cut", "", "")
	err := FlagErrorWithHints(cmd, errors.New("unknown flag: --cot"))
	if err == nil {
		t.Fatal("error = nil")
	}
	if strings.Contains(err.Error(), "did you mean") {
		t.Fatalf("error = %q, want no ambiguous suggestion", err)
	}
}
