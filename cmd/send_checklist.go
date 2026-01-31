// Package cmd provides CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	sendChecklistJSON     bool
	sendChecklistCorrect int
	sendChecklistPeer     string
	sendChecklistUsername string
)

// sendChecklistCmd represents the send-checklist command.
var sendChecklistCmd = &cobra.Command{
	Use:   "send-checklist <question> -o <option1> -o <option2> ... -c <correct_index>",
	Short: "Send a quiz (checklist) to a Telegram peer",
	Long: `Send a quiz with correct answer to a Telegram user or chat.

Provide options using -o flag. Minimum 2, maximum 10 options.
Use -c or --correct to specify the correct answer index (0-based).

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(1),
	Run:  runSendChecklist,
}

func init() {
	sendChecklistCmd.Flags().BoolVarP(&sendChecklistJSON, "json", "j", false, "Output as JSON")
	sendChecklistCmd.Flags().IntVarP(&sendChecklistCorrect, "correct", "c", 0, "Correct answer index")
	sendChecklistCmd.Flags().StringSliceP("options", "o", []string{}, "Quiz options")
	sendChecklistCmd.Flags().StringVarP(&sendChecklistPeer, "peer", "p", "", "Peer (e.g., @username)")
	sendChecklistCmd.Flags().StringVarP(&sendChecklistUsername, "username", "u", "", "Username (without @)")
	sendChecklistCmd.MarkFlagsOneRequired("peer", "username")
	sendChecklistCmd.MarkFlagsMutuallyExclusive("peer", "username")
	rootCmd.AddCommand(sendChecklistCmd)
}

func runSendChecklist(cmd *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(sendChecklistJSON)
	question := args[0]

	options, _ := cmd.Flags().GetStringSlice("options")

	if len(options) < 2 {
		fmt.Fprintf(os.Stderr, "Error: at least 2 options are required\n")
		os.Exit(1)
	}
	if len(options) > 10 {
		fmt.Fprintf(os.Stderr, "Error: maximum 10 options allowed\n")
		os.Exit(1)
	}

	if sendChecklistCorrect < 0 || sendChecklistCorrect >= len(options) {
		fmt.Fprintf(os.Stderr, "Error: correct index must be between 0 and %d\n", len(options)-1)
		os.Exit(1)
	}

	// Convert options to map format
	optionMaps := make([]map[string]string, len(options))
	for i, opt := range options {
		optionMaps[i] = map[string]string{"text": opt}
	}

	params := map[string]any{
		"question":   question,
		"options":    optionMaps,
		"quiz":       true,
		"correctIdx": sendChecklistCorrect,
	}
	if sendChecklistPeer != "" {
		params["peer"] = sendChecklistPeer
	} else {
		params["username"] = sendChecklistUsername
	}

	result := runner.CallWithParams("send_checklist", params)

	runner.PrintResult(result, func(r any) {
		rMap, _ := r.(map[string]any)
		id, _ := rMap["id"].(float64)
		peer, _ := rMap["peer"].(string)
		question, _ := rMap["question"].(string)
		fmt.Printf("Quiz sent successfully!\n")
		fmt.Printf("  Peer: @%s\n", peer)
		fmt.Printf("  Question: %s\n", question)
		fmt.Printf("  Correct: option %d\n", sendChecklistCorrect)
		fmt.Printf("  ID: %d\n", int64(id))
	})
}
