// Package send provides commands for sending messages and media.
package send

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// SendFlags holds common flags for all send commands.
//revive:disable:exported stutter
type SendFlags struct {
	To      cliutil.Recipient
	Caption string
	cmd     *cobra.Command
}

// Register registers common flags on a cobra command (with caption).
func (f *SendFlags) Register(command *cobra.Command) {
	f.cmd = command
	command.Flags().VarP(&f.To, "to", "t", "Recipient (@username, username, or chat ID)")
	command.Flags().StringVar(&f.Caption, "caption", "", "Caption")
	_ = command.MarkFlagRequired("to")
}

// RegisterOptionalTo registers common flags with --to as optional (not required).
func (f *SendFlags) RegisterOptionalTo(command *cobra.Command) {
	f.cmd = command
	command.Flags().VarP(&f.To, "to", "t", "Recipient (@username, username, or chat ID)")
	command.Flags().StringVar(&f.Caption, "caption", "", "Caption")
}

// RegisterWithoutCaption registers flags without caption option.
func (f *SendFlags) RegisterWithoutCaption(command *cobra.Command) {
	f.cmd = command
	command.Flags().VarP(&f.To, "to", "t", "Recipient (@username, username, or chat ID)")
	_ = command.MarkFlagRequired("to")
}

// AddToParams adds flags to params map.
func (f *SendFlags) AddToParams(params map[string]any) {
	f.To.AddToParams(params)
	if f.Caption != "" {
		params["caption"] = f.Caption
	}
}

// NewRunner returns a Runner from SendFlags.
func (f *SendFlags) NewRunner() *cliutil.Runner {
	return cliutil.NewRunnerFromCmd(f.cmd, false)
}

// CommandConfig holds configuration for creating a send command.
type CommandConfig struct {
	Use        string
	Short      string
	Long       string
	Args       cobra.PositionalArgs
	HasCaption bool
}

// RequirePeer exits with an error if no peer is set.
func (f *SendFlags) RequirePeer() {
	if f.To.Peer() == "" {
		fmt.Fprintln(os.Stderr, "Error: peer is required (positional or --to)")
		os.Exit(1)
	}
}

// NewCommand creates a new cobra.Command with SendFlags registered.
func (f *SendFlags) NewCommand(cfg CommandConfig, runFn func(*cobra.Command, []string)) *cobra.Command {
	command := &cobra.Command{
		Use:   cfg.Use,
		Short: cfg.Short,
		Long:  cfg.Long,
		Args:  cfg.Args,
	}

	if cfg.HasCaption {
		f.Register(command)
	} else {
		f.RegisterWithoutCaption(command)
	}
	command.Run = runFn

	return command
}
