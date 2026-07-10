package docs

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

func TestUpdateReadmeReplacesGeneratedBlocks(t *testing.T) {
	root := &cobra.Command{Use: "agent-telegram"}
	root.PersistentFlags().String("output", "", "Output format: json or ids")
	child := &cobra.Command{
		Use:     "my-info",
		Short:   "Get your profile information",
		GroupID: "auth",
		Run:     func(*cobra.Command, []string) {},
	}
	root.AddCommand(child)
	cliutil.RegisterMethod(child, "get_me")

	src := []byte(`# Test

## Commands

<!-- BEGIN GENERATED:commands -->
old commands
<!-- END GENERATED:commands -->

## Global Options

<!-- BEGIN GENERATED:global-options -->
old options
<!-- END GENERATED:global-options -->
`)

	updated, err := UpdateReadme(root, src)
	if err != nil {
		t.Fatal(err)
	}
	body := string(updated)
	for _, want := range []string{
		"Run `agent-telegram --help`",
		"| Authentication Commands | `my-info` |",
		"`--output <string>`",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("updated README missing %q:\n%s", want, body)
		}
	}
	for _, old := range []string{"old commands", "old options"} {
		if strings.Contains(body, old) {
			t.Fatalf("updated README still contains %q:\n%s", old, body)
		}
	}
}

func TestUpdateReadmeRequiresMarkers(t *testing.T) {
	root := &cobra.Command{Use: "agent-telegram"}
	_, err := UpdateReadme(root, []byte("# Test\n"))
	if err == nil {
		t.Fatal("UpdateReadme should fail when generated markers are missing")
	}
}

func TestGenerateLLMMarkdownMentionsProjectAwareSkillOnboarding(t *testing.T) {
	root := &cobra.Command{Use: "agent-telegram"}
	body := GenerateLLMMarkdown(root)
	for _, want := range []string{
		"existing project `.agents/skills`",
		"global `$HOME/.agents/skills` installation requires consent",
		"agent-telegram skills install agent-telegram",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("generated docs missing %q:\n%s", want, body)
		}
	}
}
