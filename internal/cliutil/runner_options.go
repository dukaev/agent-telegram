package cliutil

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type runnerFlagOptions struct {
	socketPath   string
	quiet        bool
	format       OutputFormat
	fields       []string
	filters      []string
	verbosity    string
	maxItems     int
	maxTextChars int
	include      []string
	omit         []string
	summary      bool
	receipt      bool
	dryRun       bool
	agentMode    bool
	runID        string
}

func runnerFlagOptionsFromCmd(cmd *cobra.Command) runnerFlagOptions {
	opts := runnerFlagOptions{
		socketPath: flagString(cmd, "socket"),
		quiet:      flagBool(cmd, "quiet"),
		format:     ParseOutputFormat(flagString(cmd, "output")),
		fields:     flagStringSlice(cmd, "fields"),
		filters:    flagStringSlice(cmd, "filter"),
		verbosity:  flagString(cmd, "verbosity"),
		include:    flagStringSlice(cmd, "include"),
		omit:       flagStringSlice(cmd, "omit"),
		summary:    flagBool(cmd, "summary"),
		receipt:    flagBool(cmd, "receipt"),
		dryRun:     flagBool(cmd, "dry-run"),
		agentMode:  flagBool(cmd, "agent"),
		runID:      flagString(cmd, "run-id"),
	}
	opts.maxItems, _ = cmd.Flags().GetInt("max-items")
	opts.maxTextChars, _ = cmd.Flags().GetInt("max-text-chars")
	opts.applyAgentDefaults(cmd)
	return opts
}

func (o *runnerFlagOptions) applyAgentDefaults(cmd *cobra.Command) {
	if !o.agentMode {
		return
	}
	o.format = OutputJSON
	o.receipt = true
	o.quiet = true
	if !flagChanged(cmd, "verbosity") {
		o.verbosity = string(VerbosityCompact)
	}
	if o.maxItems <= 0 {
		o.maxItems = 8
	}
	if o.maxTextChars <= 0 {
		o.maxTextChars = 180
	}
}

func (o runnerFlagOptions) filterExpressions() FilterExpressions {
	if len(o.filters) == 0 {
		return nil
	}
	filterExprs, err := ParseFilterExpressions(o.filters)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		Exit(1)
	}
	return filterExprs
}

func (o runnerFlagOptions) outputBudget() OutputBudgetOptions {
	return OutputBudgetOptions{
		Verbosity:    ParseVerbosity(o.verbosity),
		MaxItems:     o.maxItems,
		MaxTextChars: o.maxTextChars,
		Include:      o.include,
		Omit:         o.omit,
		Summary:      o.summary,
	}
}

func flagString(cmd *cobra.Command, name string) string {
	value, _ := cmd.Flags().GetString(name)
	return value
}

func flagBool(cmd *cobra.Command, name string) bool {
	value, _ := cmd.Flags().GetBool(name)
	return value
}

func flagStringSlice(cmd *cobra.Command, name string) []string {
	value, _ := cmd.Flags().GetStringSlice(name)
	return value
}
