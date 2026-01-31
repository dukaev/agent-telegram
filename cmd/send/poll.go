// Package send provides commands for sending messages and media.
package send

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// sendPollAnonymous enables anonymous poll mode.
	sendPollAnonymous bool
	// sendPollQuiz enables quiz mode.
	sendPollQuiz bool
	// sendPollCorrect is the correct answer index (for quiz).
	sendPollCorrect int
)

// PollCmd represents the send-poll command.
var PollCmd = &cobra.Command{
	Use:   "send-poll <question> -o <option1> -o <option2> ...",
	Short: "Send a poll to a Telegram peer",
	Long: `Send a poll to a Telegram user or chat.

Provide options using -o flag. Minimum 2, maximum 10 options.

For quiz mode, use --quiz and --correct <index> (0-based).

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
	Args: cobra.ExactArgs(1),
}

var sendPollFlags SendFlags

// AddPollCommand adds the poll command to the root command.
func AddPollCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(PollCmd)

	PollCmd.Flags().BoolVar(&sendPollAnonymous, "anonymous", false, "Anonymous poll")
	PollCmd.Flags().BoolVar(&sendPollQuiz, "quiz", false, "Quiz mode")
	PollCmd.Flags().IntVar(&sendPollCorrect, "correct", 0, "Correct answer index (for quiz)")
	PollCmd.Flags().StringSliceP("options", "o", []string{}, "Poll options")

	sendPollFlags.RegisterWithoutCaption(PollCmd)

	PollCmd.Run = runPoll
}

func runPoll(command *cobra.Command, args []string) {
	runner := sendPollFlags.NewRunner()
	question := args[0]

	options, _ := command.Flags().GetStringSlice("options")

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

	params := buildPollParams(question, options)
	result := runner.CallWithParams("send_poll", params)

	runner.PrintResult(result, func(r any) {
		printPollResult(r, question)
	})
}

func buildPollParams(question string, options []string) map[string]any {
	optionMaps := make([]map[string]string, len(options))
	for i, opt := range options {
		optionMaps[i] = map[string]string{"text": opt}
	}

	params := map[string]any{
		"question":   question,
		"options":    optionMaps,
		"anonymous":  sendPollAnonymous,
		"quiz":       sendPollQuiz,
		"correctIdx": sendPollCorrect,
	}
	sendPollFlags.AddToParams(params)
	return params
}

func printPollResult(r any, question string) {
	rMap, _ := r.(map[string]any)
	id, _ := rMap["id"].(float64)
	fmt.Printf("Poll sent successfully!\n")
	fmt.Printf("  Question: %s\n", question)
	fmt.Printf("  ID: %d\n", int64(id))
}
