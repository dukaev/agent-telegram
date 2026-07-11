package cliutil

import (
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const firstArgPeerAnnotation = "agent-telegram.io/first-arg-peer"

var negativePeerPattern = regexp.MustCompile(`^-[0-9]+$`)

// MarkFirstArgPeer opts a command into negative positional peer normalization.
func MarkFirstArgPeer(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[firstArgPeerAnnotation] = "true"
}

// AcceptsFirstArgPeer reports whether a command's first positional argument is a peer.
func AcceptsFirstArgPeer(cmd *cobra.Command) bool {
	return cmd != nil && cmd.Annotations[firstArgPeerAnnotation] == "true"
}

// NormalizeNegativePeerArgs rewrites a strict negative decimal in the first
// positional peer slot into pflag's unambiguous --to=<peer> form. The input
// slice and Cobra flag parsing state are left untouched.
func NormalizeNegativePeerArgs(root *cobra.Command, args []string) []string {
	normalized := append([]string(nil), args...)
	cmd, tail := commandForArgs(root, normalized)
	if !AcceptsFirstArgPeer(cmd) {
		return normalized
	}

	for i := tail; i < len(normalized); i++ {
		arg := normalized[i]
		if arg == "--" {
			return normalized
		}
		if known, consumesNext := knownFlagToken(cmd, arg); known {
			if consumesNext && i+1 < len(normalized) {
				i++
			}
			continue
		}
		if negativePeerPattern.MatchString(arg) {
			normalized[i] = "--to=" + arg
			return normalized
		}
		if strings.HasPrefix(arg, "-") && arg != "-" {
			// Preserve unknown flags for Cobra to diagnose.
			continue
		}
		// Only the first positional argument can be a peer.
		return normalized
	}
	return normalized
}

// commandForArgs finds the deepest command and the beginning of its argument
// tail without calling Cobra Find, which treats negative decimals as flags.
func commandForArgs(root *cobra.Command, args []string) (*cobra.Command, int) {
	cmd := root
	tail := 0
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			return cmd, tail
		}
		if known, consumesNext := knownFlagToken(cmd, arg); known {
			if consumesNext && i+1 < len(args) {
				i++
			}
			continue
		}
		if child := matchingChild(cmd, arg); child != nil {
			cmd = child
			tail = i + 1
			continue
		}
		if strings.HasPrefix(arg, "-") && arg != "-" {
			continue
		}
		return cmd, tail
	}
	return cmd, tail
}

func matchingChild(cmd *cobra.Command, arg string) *cobra.Command {
	for _, child := range cmd.Commands() {
		if child.Name() == arg {
			return child
		}
		for _, alias := range child.Aliases {
			if alias == arg {
				return child
			}
		}
	}
	return nil
}

func knownFlagToken(cmd *cobra.Command, arg string) (known, consumesNext bool) {
	if strings.HasPrefix(arg, "--") && len(arg) > 2 {
		name, _, hasValue := strings.Cut(arg[2:], "=")
		flag := lookupFlag(cmd, name)
		if flag == nil {
			return false, false
		}
		return true, !hasValue && flag.NoOptDefVal == ""
	}
	if !strings.HasPrefix(arg, "-") || len(arg) < 2 || arg == "--" {
		return false, false
	}

	shorthands := arg[1:]
	for i := 0; i < len(shorthands); i++ {
		flag := lookupShorthand(cmd, shorthands[i:i+1])
		if flag == nil {
			return false, false
		}
		if flag.NoOptDefVal == "" {
			// Any remaining bytes are the value of this shorthand.
			return true, i == len(shorthands)-1
		}
	}
	return true, false
}

func lookupFlag(cmd *cobra.Command, name string) *pflag.Flag {
	if flag := cmd.Flags().Lookup(name); flag != nil {
		return flag
	}
	for parent := cmd; parent != nil; parent = parent.Parent() {
		if flag := parent.PersistentFlags().Lookup(name); flag != nil {
			return flag
		}
	}
	return nil
}

func lookupShorthand(cmd *cobra.Command, shorthand string) *pflag.Flag {
	if flag := cmd.Flags().ShorthandLookup(shorthand); flag != nil {
		return flag
	}
	for parent := cmd; parent != nil; parent = parent.Parent() {
		if flag := parent.PersistentFlags().ShorthandLookup(shorthand); flag != nil {
			return flag
		}
	}
	return nil
}
