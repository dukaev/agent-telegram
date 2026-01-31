// Package cmd provides CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	sendPollJSON      bool
	sendPollAnonymous bool
	sendPollQuiz      bool
	sendPollCorrect   int
)

// sendPollCmd represents the send-poll command.
var sendPollCmd = &cobra.Command{
	Use:   "send-poll @peer <question> -o <option1> -o <option2> ...",
	Short: "Send a poll to a Telegram peer",
	Long: `Send a poll to a Telegram user or chat.

Provide options using -o flag. Minimum 2, maximum 10 options.

For quiz mode, use --quiz and --correct <index> (0-based).

Example: agent-telegram send-poll @user "Best framework?" -o "React" -o "Vue" -o "Angular"`,
	Args: cobra.ExactArgs(2),
	Run:  runSendPoll,
}

func init() {
	sendPollCmd.Flags().BoolVarP(&sendPollJSON, "json", "j", false, "Output as JSON")
	sendPollCmd.Flags().BoolVarP(&sendPollAnonymous, "anonymous", "a", false, "Anonymous poll")
	sendPollCmd.Flags().BoolVar(&sendPollQuiz, "quiz", false, "Quiz mode")
	sendPollCmd.Flags().IntVar(&sendPollCorrect, "correct", 0, "Correct answer index (for quiz)")
	sendPollCmd.Flags().StringSliceP("options", "o", []string{}, "Poll options")
	rootCmd.AddCommand(sendPollCmd)
}

func runSendPoll(cmd *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(sendPollJSON)
	peer := args[0]
	question := args[1]

	options, _ := cmd.Flags().GetStringSlice("options")

	if len(options) < 2 {
		fmt.Fprintf(os.Stderr, "Error: at least 2 options are required\n")
		os.Exit(1)
	}
	if len(options) > 10 {
		fmt.Fprintf(os.Stderr, "Error: maximum 10 options allowed\n")
		os.Exit(1)
	}

	if sendPollQuiz && (sendPollCorrect < 0 || sendPollCorrect >= len(options)) {
		fmt.Fprintf(os.Stderr, "Error: correct index must be between 0 and %d\n", len(options)-1)
		os.Exit(1)
	}

	// Convert options to map format
	optionMaps := make([]map[string]string, len(options))
	for i, opt := range options {
		optionMaps[i] = map[string]string{"text": opt}
	}

	result := runner.CallWithParams("send_poll", map[string]any{
		"peer":       peer,
		"question":   question,
		"options":    optionMaps,
		"anonymous":  sendPollAnonymous,
		"quiz":       sendPollQuiz,
		"correctIdx": sendPollCorrect,
	})

	runner.PrintResult(result, func(r any) {
		rMap, _ := r.(map[string]any)
		id, _ := rMap["id"].(float64)
		peer, _ := rMap["peer"].(string)
		question, _ := rMap["question"].(string)
		fmt.Printf("Poll sent successfully!\n")
		fmt.Printf("  Peer: @%s\n", peer)
		fmt.Printf("  Question: %s\n", question)
		fmt.Printf("  ID: %d\n", int64(id))
	})
}
