// Package contact provides commands for managing contacts.
package contact

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	listSearchQuery string
	listLimit       int
	listOffset      int
)

// ListContactsCmd represents the contact list command.
var ListContactsCmd = &cobra.Command{
	Use:     "list",
	Short:   "List contacts from your Telegram account",
	Long: `List all contacts from your Telegram account with optional search filter.

Search matches against:
  - First name
  - Last name
  - Username
  - Phone number

Example:
  agent-telegram contact list
  agent-telegram contact list --search john
  agent-telegram contact list --search "@john" --limit 20`,
	Args: cobra.NoArgs,
}

// AddListContactsCommand adds the list contacts command to the root command.
func AddListContactsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ListContactsCmd)

	ListContactsCmd.Flags().StringVarP(&listSearchQuery, "search", "Q", "", "Search query to filter contacts")
	ListContactsCmd.Flags().IntVarP(&listLimit, "limit", "l", cliutil.DefaultLimitLarge, "Max contacts (max 100)")
	ListContactsCmd.Flags().IntVarP(&listOffset, "offset", "o", 0, "Offset for pagination")

	ListContactsCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(ListContactsCmd, false)
		params := buildListParams()
		result := runner.CallWithParams("get_contacts", params)
		runner.PrintResult(result, printContacts)
	}
}

func buildListParams() map[string]any {
	pag := cliutil.NewPagination(listLimit, listOffset, cliutil.PaginationConfig{
		MaxLimit: cliutil.MaxLimitStandard,
	})

	params := map[string]any{}
	pag.ToParams(params, true)
	if listSearchQuery != "" {
		params["query"] = listSearchQuery
	}
	return params
}

func printContacts(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to get contacts")
		return
	}

	printContactsHeader(r)
	contacts, _ := r["contacts"].([]any)
	for _, c := range contacts {
		printContact(c)
	}
}

func printContactsHeader(r map[string]any) {
	query, _ := r["query"].(string)
	count, _ := r["count"].(float64)
	if query != "" {
		fmt.Fprintf(os.Stderr, "Found %d contacts matching '%s':\n", int(count), query)
	} else {
		fmt.Fprintf(os.Stderr, "Contacts (%d):\n", int(count))
	}
}

func printContact(c any) {
	contact, ok := c.(map[string]any)
	if !ok {
		return
	}
	name := buildContactName(contact)
	line := formatContactLine(name, contact)
	fmt.Fprintln(os.Stderr, line)
}

func buildContactName(contact map[string]any) string {
	firstName, _ := contact["firstName"].(string)
	lastName, _ := contact["lastName"].(string)
	name := firstName
	if lastName != "" {
		if name != "" {
			name += " " + lastName
		} else {
			name = lastName
		}
	}
	if name == "" {
		name = "Unknown"
	}
	return name
}

func formatContactLine(name string, contact map[string]any) string {
	username, _ := contact["username"].(string)
	phone, _ := contact["phone"].(string)
	peer, _ := contact["peer"].(string)
	bot, _ := contact["bot"].(bool)
	verified, _ := contact["verified"].(bool)

	line := fmt.Sprintf("  - %s", name)
	if username != "" {
		line += fmt.Sprintf(" (@%s)", username)
	}
	if peer != "" {
		line += fmt.Sprintf(" [%s]", peer)
	}
	if bot {
		line += " 🤖"
	}
	if verified {
		line += " ✓"
	}
	if phone != "" {
		line += fmt.Sprintf(" - %s", phone)
	}
	return line
}
