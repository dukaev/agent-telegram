package contact

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestAddContactCommandRegistersExpectedSurface(t *testing.T) {
	root := &cobra.Command{Use: "root"}

	AddContactCommand(root)

	contactCmd := childCommand(root, "contact")
	if contactCmd == nil {
		t.Fatal("contact command was not registered")
	}
	for _, name := range []string{"add", "delete", "list"} {
		if childCommand(contactCmd, name) == nil {
			t.Fatalf("contact subcommand %q was not registered", name)
		}
	}
	if AddContactCmd.Flags().Lookup("phone") == nil || AddContactCmd.Flags().Lookup("first-name") == nil {
		t.Fatal("expected add contact flags")
	}
	if ListContactsCmd.Flags().Lookup("search") == nil || ListContactsCmd.Flags().Lookup("limit") == nil {
		t.Fatal("expected list contact flags")
	}
}

func TestBuildListParams(t *testing.T) {
	listSearchQuery = "john"
	listLimit = 500
	listOffset = -10
	t.Cleanup(func() {
		listSearchQuery = ""
		listLimit = 0
		listOffset = 0
	})

	params := buildListParams()
	if params["query"] != "john" {
		t.Fatalf("query = %v, want john", params["query"])
	}
	if params["limit"] != 100 {
		t.Fatalf("limit = %v, want capped 100", params["limit"])
	}
	if params["offset"] != 0 {
		t.Fatalf("offset = %v, want normalized 0", params["offset"])
	}
}

func TestContactFormatting(t *testing.T) {
	contact := map[string]any{
		"firstName": "Ada",
		"lastName":  "Lovelace",
		"username":  "ada",
		"phone":     "+123",
		"peer":      "@ada",
		"bot":       true,
		"verified":  true,
	}
	if got := buildContactName(contact); got != "Ada Lovelace" {
		t.Fatalf("name = %q, want full name", got)
	}
	line := formatContactLine("Ada Lovelace", contact)
	for _, want := range []string{"Ada Lovelace", "@ada", "@ada", "+123"} {
		if !strings.Contains(line, want) {
			t.Fatalf("line %q missing %q", line, want)
		}
	}
	if got := buildContactName(map[string]any{}); got != "Unknown" {
		t.Fatalf("empty name = %q, want Unknown", got)
	}
}

func TestPrintContactsAndAddResult(t *testing.T) {
	output := captureStderr(t, func() {
		printContacts(map[string]any{
			"count": float64(1),
			"query": "ada",
			"contacts": []any{map[string]any{
				"firstName": "Ada",
				"username":  "ada",
			}},
		})
		printContact("ignored")
		printContactResult(map[string]any{
			"contact": map[string]any{
				"firstName": "Grace",
				"lastName":  "Hopper",
				"username":  "grace",
				"phone":     "+456",
				"peer":      "@grace",
			},
		})
		printContactResult("ok")
	})

	for _, want := range []string{"Found 1 contacts", "Ada", "Added contact: Grace Hopper", "Contact added successfully"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}

	fallback := captureStderr(t, func() { printContacts("bad") })
	if !strings.Contains(fallback, "Failed to get contacts") {
		t.Fatalf("fallback output = %q", fallback)
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

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	defer func() { os.Stderr = old }()

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return <-done
}
