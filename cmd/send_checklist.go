// Package cmd provides CLI commands.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

var (
	sendChecklistJSON bool
	sendChecklistCorrect int
)

// sendChecklistCmd represents the send-checklist command.
var sendChecklistCmd = &cobra.Command{
	Use:   "send-checklist @peer <question> -o <option1> -o <option2> ... -c <correct_index>",
	Short: "Send a quiz (checklist) to a Telegram peer",
	Long: `Send a quiz with correct answer to a Telegram user or chat.

Provide options using -o flag. Minimum 2, maximum 10 options.
Use -c or --correct to specify the correct answer index (0-based).

Example: agent-telegram send-checklist @user "What is 2+2?" -o "3" -o "4" -o "5" -c 1`,
	Args: cobra.ExactArgs(2),
	Run:  runSendChecklist,
}

func init() {
	sendChecklistCmd.Flags().BoolVarP(&sendChecklistJSON, "json", "j", false, "Output as JSON")
	sendChecklistCmd.Flags().IntVarP(&sendChecklistCorrect, "correct", "c", 0, "Correct answer index")
	sendChecklistCmd.Flags().StringSliceP("options", "o", []string{}, "Quiz options")
	rootCmd.AddCommand(sendChecklistCmd)
}

func runSendChecklist(cmd *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
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

	if sendChecklistCorrect < 0 || sendChecklistCorrect >= len(options) {
		fmt.Fprintf(os.Stderr, "Error: correct index must be between 0 and %d\n", len(options)-1)
		os.Exit(1)
	}

	// Convert options to map format
	optionMaps := make([]map[string]string, len(options))
	for i, opt := range options {
		optionMaps[i] = map[string]string{"text": opt}
	}

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("send_checklist", map[string]any{
		"peer":       peer,
		"question":   question,
		"options":    optionMaps,
		"quiz":       true,
		"correctIdx": sendChecklistCorrect,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	if sendChecklistJSON {
		printSendChecklistJSON(result)
	} else {
		printSendChecklistResult(result)
	}
}

// printSendChecklistJSON prints the result as JSON.
func printSendChecklistJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printSendChecklistResult prints the result in a human-readable format.
func printSendChecklistResult(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	id, _ := r["id"].(float64)
	peer, _ := r["peer"].(string)
	question, _ := r["question"].(string)

	fmt.Printf("Quiz sent successfully!\n")
	fmt.Printf("  Peer: @%s\n", peer)
	fmt.Printf("  Question: %s\n", question)
	fmt.Printf("  Correct: option %d\n", sendChecklistCorrect)
	fmt.Printf("  ID: %d\n", int64(id))
}
