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

	if sendPollQuiz && (sendPollCorrect < 0 || sendPollCorrect >= len(options)) {
		fmt.Fprintf(os.Stderr, "Error: correct index must be between 0 and %d\n", len(options)-1)
		os.Exit(1)
	}

	// Convert options to map format
	optionMaps := make([]map[string]string, len(options))
	for i, opt := range options {
		optionMaps[i] = map[string]string{"text": opt}
	}

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("send_poll", map[string]any{
		"peer":       peer,
		"question":   question,
		"options":    optionMaps,
		"anonymous":  sendPollAnonymous,
		"quiz":       sendPollQuiz,
		"correctIdx": sendPollCorrect,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	if sendPollJSON {
		printSendPollJSON(result)
	} else {
		printSendPollResult(result)
	}
}

// printSendPollJSON prints the result as JSON.
func printSendPollJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printSendPollResult prints the result in a human-readable format.
func printSendPollResult(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	id, _ := r["id"].(float64)
	peer, _ := r["peer"].(string)
	question, _ := r["question"].(string)

	fmt.Printf("Poll sent successfully!\n")
	fmt.Printf("  Peer: @%s\n", peer)
	fmt.Printf("  Question: %s\n", question)
	fmt.Printf("  ID: %d\n", int64(id))
}
