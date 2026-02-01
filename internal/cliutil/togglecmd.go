package cliutil

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ToggleCommandConfig defines configuration for toggle commands.
type ToggleCommandConfig struct {
	Use           string
	Short         string
	Long          string
	EnableMethod  string // Method to call when enabling (if empty, uses SingleMethod with disable=false)
	DisableMethod string // Method to call when disabling (if empty, uses SingleMethod with disable=true)
	SingleMethod  string // Single method with disable param (used if EnableMethod is empty)
	EnableMsg     string // Success message when enabling
	DisableMsg    string // Success message when disabling
}

// resolveToggleMethod returns the method and success message based on config and disable flag.
func resolveToggleMethod(cfg ToggleCommandConfig, disable bool) (method, successMsg string) {
	switch {
	case cfg.SingleMethod != "":
		method = cfg.SingleMethod
	case disable:
		method = cfg.DisableMethod
	default:
		method = cfg.EnableMethod
	}

	switch disable {
	case true:
		successMsg = cfg.DisableMsg
	default:
		successMsg = cfg.EnableMsg
	}
	return method, successMsg
}

// NewToggleCommand creates a cobra command for toggle operations.
func NewToggleCommand(cfg ToggleCommandConfig) *cobra.Command {
	var to Recipient
	var disable bool

	cmd := &cobra.Command{
		Use:   cfg.Use,
		Short: cfg.Short,
		Long:  cfg.Long,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			runner := NewRunnerFromCmd(cmd, false)
			params := map[string]any{}
			to.AddToParams(params)

			var method string
			var successMsg string

			method, successMsg = resolveToggleMethod(cfg, disable)
			if cfg.SingleMethod != "" {
				params["disable"] = disable
			}

			result := runner.CallWithParams(method, params)
			runner.PrintResult(result, func(result any) {
				if runner.IsQuiet() {
					return
				}
				r, ok := result.(map[string]any)
				if !ok {
					fmt.Fprintf(os.Stderr, "%s\n", successMsg)
					return
				}
				peer := ExtractString(r, "peer")
				fmt.Fprintf(os.Stderr, "%s\n", successMsg)
				if peer != "" {
					fmt.Fprintf(os.Stderr, "  Peer: %s\n", peer)
				}
			})
		},
	}

	cmd.Flags().VarP(&to, "to", "t", "Recipient (@username, username, or chat ID)")
	cmd.Flags().BoolVarP(&disable, "disable", "d", false, "Disable/reverse the action")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}
