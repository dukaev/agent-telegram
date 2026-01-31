// Package cmd provides common structures for send commands.
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// Recipient represents a message recipient.
// Accepts: @username, username (without @), or chat ID (numeric).
type Recipient struct {
	value string
}

func (r *Recipient) String() string {
	return r.value
}

func (r *Recipient) Set(s string) error {
	if s == "" {
		return fmt.Errorf("recipient cannot be empty")
	}
	r.value = s
	return nil
}

func (r *Recipient) Type() string {
	return "recipient"
}

// Peer returns normalized peer for API.
// @user → @user
// username → @username
// 123456789 → 123456789 (chat ID)
func (r *Recipient) Peer() string {
	if r.value == "" {
		return ""
	}
	if strings.HasPrefix(r.value, "@") {
		return r.value
	}
	if r.value[0] >= '0' && r.value[0] <= '9' {
		return r.value
	}
	return "@" + r.value
}

// AddToParams adds normalized peer to parameters.
func (r *Recipient) AddToParams(params map[string]any) {
	params["peer"] = r.Peer()
}

// SendFlags holds common flags for all send commands.
type SendFlags struct {
	JSON    bool
	To      Recipient
	Caption string
}

// Register registers common flags on a cobra command (with caption).
func (f *SendFlags) Register(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&f.JSON, "json", "j", false, "Output as JSON")
	cmd.Flags().VarP(&f.To, "to", "t", "Recipient (@username, username, or chat ID)")
	cmd.Flags().StringVar(&f.Caption, "caption", "", "Caption")
	cmd.MarkFlagRequired("to")
}

// RegisterWithoutCaption registers flags without caption option.
func (f *SendFlags) RegisterWithoutCaption(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&f.JSON, "json", "j", false, "Output as JSON")
	cmd.Flags().VarP(&f.To, "to", "t", "Recipient (@username, username, or chat ID)")
	cmd.MarkFlagRequired("to")
}

// AddToParams adds flags to params map.
func (f *SendFlags) AddToParams(params map[string]any) {
	f.To.AddToParams(params)
	if f.Caption != "" {
		params["caption"] = f.Caption
	}
}

// NewRunner returns a Runner from SendFlags.
func (f *SendFlags) NewRunner() *Runner {
	return NewRunnerFromRoot(f.JSON)
}

// CommandConfig holds configuration for creating a send command.
type CommandConfig struct {
	Use     string
	Short   string
	Long    string
	Args    cobra.PositionalArgs
	Run     func(*cobra.Command, []string)
	HasCaption bool
}

// NewCommand creates a new cobra.Command with SendFlags registered.
func (f *SendFlags) NewCommand(cfg CommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cfg.Use,
		Short: cfg.Short,
		Long:  cfg.Long,
		Args:  cfg.Args,
		Run:   cfg.Run,
	}

	if cfg.HasCaption {
		f.Register(cmd)
	} else {
		f.RegisterWithoutCaption(cmd)
	}

	return cmd
}
