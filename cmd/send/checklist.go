// Package send provides commands for sending messages and media.
package send

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// sendChecklistCorrect is the correct answer index.
	sendChecklistCorrect int
)

// ChecklistCmd represents the send-checklist command.
var ChecklistCmd = &cobra.Command{
	GroupID: "messaging",
	Use:   "send-checklist <question> -o <option1> -o <option2> ... -c <correct_index>",
	Short: "Send a quiz (checklist) to a Telegram peer",
	Long: `Send a quiz with correct answer to a Telegram user or chat.

Provide options using -o flag. Minimum 2, maximum 10 options.
Use -c or --correct to specify the correct answer index (0-based).

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
	Args: cobra.ExactArgs(1),
}

var sendChecklistFlags SendFlags

// AddChecklistCommand adds the checklist command to the root command.
func AddChecklistCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ChecklistCmd)

	ChecklistCmd.Flags().IntVarP(&sendChecklistCorrect, "correct", "c", 0, "Correct answer index")
	ChecklistCmd.Flags().StringSliceP("options", "o", []string{}, "Quiz options")

	sendChecklistFlags.RegisterWithoutCaption(ChecklistCmd)

	ChecklistCmd.Run = runChecklist
}

func runChecklist(command *cobra.Command, args []string) {
	runner := sendChecklistFlags.NewRunner()
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

	if sendChecklistCorrect < 0 || sendChecklistCorrect >= len(options) {
		fmt.Fprintf(os.Stderr, "Error: correct index must be between 0 and %d\n", len(options)-1)
		os.Exit(1)
	}

	params := buildChecklistParams(question, options)
	result := runner.CallWithParams("send_checklist", params)

	runner.PrintResult(result, func(r any) {
		printChecklistResult(r, question)
	})
}

func buildChecklistParams(question string, options []string) map[string]any {
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
	sendChecklistFlags.AddToParams(params)
	return params
}

func printChecklistResult(r any, question string) {
	rMap, _ := r.(map[string]any)
	id, _ := rMap["id"].(float64)
	fmt.Printf("Quiz sent successfully!\n")
	fmt.Printf("  Question: %s\n", question)
	fmt.Printf("  Correct: option %d\n", sendChecklistCorrect)
	fmt.Printf("  ID: %d\n", int64(id))
}
