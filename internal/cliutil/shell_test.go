package cliutil

import "testing"

func TestShellArg(t *testing.T) {
	tests := map[string]string{
		"-5424738551": "-5424738551",
		"@bot":        "@bot",
		"two words":   "'two words'",
		"it's":        "'it'\\''s'",
		"$HOME":       "'$HOME'",
		"`whoami`":    "'`whoami`'",
	}
	for input, want := range tests {
		if got := ShellArg(input); got != want {
			t.Errorf("ShellArg(%q) = %q, want %q", input, got, want)
		}
	}
}
