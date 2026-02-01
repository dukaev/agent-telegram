// Package contact provides commands for managing contacts.
package contact

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	listSearchQuery string
	listLimit       int
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

	ListContactsCmd.Flags().StringVarP(&listSearchQuery, "search", "s", "", "Search query to filter contacts")
	ListContactsCmd.Flags().IntVarP(&listLimit, "limit", "l", 50, "Maximum number of contacts to return")

	ListContactsCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(ListContactsCmd, false)
		params := map[string]any{}

		if listSearchQuery != "" {
			params["query"] = listSearchQuery
		}
		if listLimit > 0 {
			params["limit"] = listLimit
		}

		result := runner.CallWithParams("get_contacts", params)
		runner.PrintResult(result, func(result any) {
			r, ok := result.(map[string]any)
			if !ok {
				fmt.Println("Failed to get contacts")
				return
			}

			query, _ := r["query"].(string)
			count, _ := r["count"].(float64)

			if query != "" {
				fmt.Printf("Found %d contacts matching '%s':\n", int(count), query)
			} else {
				fmt.Printf("Contacts (%d):\n", int(count))
			}

			contacts, _ := r["contacts"].([]any)
			for _, c := range contacts {
				contact, ok := c.(map[string]any)
				if !ok {
					continue
				}

				firstName, _ := contact["firstName"].(string)
				lastName, _ := contact["lastName"].(string)
				username, _ := contact["username"].(string)
				phone, _ := contact["phone"].(string)
				peer, _ := contact["peer"].(string)
				bot, _ := contact["bot"].(bool)
				verified, _ := contact["verified"].(bool)

				// Build display name
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

				// Format output line
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
				fmt.Println(line)
			}
		})
	}
}
