package cliutil

import (
	"slices"
	"testing"

	"github.com/spf13/cobra"
)

func TestNormalizeNegativePeerArgs(t *testing.T) {
	root := &cobra.Command{Use: "agent-telegram"}
	root.PersistentFlags().String("run-id", "", "")
	root.PersistentFlags().StringP("socket", "s", "", "")
	root.PersistentFlags().BoolP("quiet", "q", false, "")
	step := &cobra.Command{Use: "step [peer]", Aliases: []string{"advance"}}
	step.Flags().String("send", "", "")
	step.Flags().StringP("to", "t", "", "")
	step.Flags().Int64("after-id", 0, "")
	MarkFirstArgPeer(step)
	bot := &cobra.Command{Use: "bot", Aliases: []string{"flow"}}
	bot.AddCommand(step)
	root.AddCommand(bot)

	unannotated := &cobra.Command{Use: "press <message-id> <button-index>"}
	bot.AddCommand(unannotated)

	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "plain", in: []string{"bot", "step", "-5424738551", "--send", "/start"}, want: []string{"bot", "step", "--to=-5424738551", "--send", "/start"}},
		{name: "global flag", in: []string{"--run-id", "run-1", "bot", "step", "-5424738551"}, want: []string{"--run-id", "run-1", "bot", "step", "--to=-5424738551"}},
		{name: "global shorthand value", in: []string{"-s", "agent.sock", "bot", "step", "-5424738551"}, want: []string{"-s", "agent.sock", "bot", "step", "--to=-5424738551"}},
		{name: "global boolean shorthand", in: []string{"-q", "bot", "step", "-5424738551"}, want: []string{"-q", "bot", "step", "--to=-5424738551"}},
		{name: "command aliases", in: []string{"flow", "advance", "-5424738551"}, want: []string{"flow", "advance", "--to=-5424738551"}},
		{name: "explicit to", in: []string{"bot", "step", "--to=-5424738551"}, want: []string{"bot", "step", "--to=-5424738551"}},
		{name: "explicit to separate value", in: []string{"bot", "step", "--to", "-5424738551"}, want: []string{"bot", "step", "--to", "-5424738551"}},
		{name: "explicit shorthand to", in: []string{"bot", "step", "-t", "-5424738551"}, want: []string{"bot", "step", "-t", "-5424738551"}},
		{name: "separator", in: []string{"bot", "step", "--", "-5424738551"}, want: []string{"bot", "step", "--", "-5424738551"}},
		{name: "not decimal", in: []string{"bot", "step", "-abc"}, want: []string{"bot", "step", "-abc"}},
		{name: "flag negative value", in: []string{"bot", "step", "--after-id", "-5", "-5424738551"}, want: []string{"bot", "step", "--after-id", "-5", "--to=-5424738551"}},
		{name: "flags after peer", in: []string{"bot", "step", "-5424738551", "--send", "hello"}, want: []string{"bot", "step", "--to=-5424738551", "--send", "hello"}},
		{name: "only first positional qualifies", in: []string{"bot", "step", "@bot", "-5"}, want: []string{"bot", "step", "@bot", "-5"}},
		{name: "negative non-peer values untouched", in: []string{"bot", "press", "-5", "-1"}, want: []string{"bot", "press", "-5", "-1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := slices.Clone(tt.in)
			if got := NormalizeNegativePeerArgs(root, tt.in); !slices.Equal(got, tt.want) {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
			if !slices.Equal(tt.in, original) {
				t.Fatalf("input mutated: got %q, want %q", tt.in, original)
			}
		})
	}
}

func TestFirstArgPeerAnnotation(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	if AcceptsFirstArgPeer(cmd) {
		t.Fatal("unmarked command accepts a first positional peer")
	}
	MarkFirstArgPeer(cmd)
	if !AcceptsFirstArgPeer(cmd) {
		t.Fatal("marked command does not accept a first positional peer")
	}
}
