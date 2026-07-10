package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootLandingPrioritizesAuthentication(t *testing.T) {
	var output bytes.Buffer
	command := &cobra.Command{}
	command.SetOut(&output)

	if err := runRoot(command, nil); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	authIndex := strings.Index(text, "agent-telegram auth")
	afterSignInIndex := strings.Index(text, "After sign-in")
	if authIndex < 0 {
		t.Fatalf("root landing does not offer web auth:\n%s", text)
	}
	if afterSignInIndex < 0 || authIndex > afterSignInIndex {
		t.Fatalf("authentication must appear before post-login commands:\n%s", text)
	}
	for _, expected := range []string{
		"Sign in with a QR code in your browser",
		"agent-telegram --help",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("root landing is missing %q:\n%s", expected, text)
		}
	}
}
