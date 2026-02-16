package cliutil

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// FlagType defines the type of flag.
type FlagType int

// Flag types for command flag definitions.
const (
	FlagString      FlagType = iota // String flag
	FlagBool                        // Boolean flag
	FlagInt                         // Integer flag
	FlagStringSlice                 // String slice flag
)

// Flag defines a command flag.
type Flag struct {
	Name      string
	Short     string
	Usage     string
	Required  bool
	Type      FlagType // Default: FlagString
	Default   any      // Default value (type must match Type)
	ParamName string   // API parameter name (if different from Name)
}

// SimpleCommandDef defines a simple RPC command.
type SimpleCommandDef struct {
	Use     string
	Short   string
	Long    string
	Method  string
	Flags   []Flag
	Success string // Message to print on success
}

// flagValues holds all flag value pointers.
type flagValues struct {
	strings      map[string]*string
	bools        map[string]*bool
	ints         map[string]*int
	stringSlices map[string]*[]string
}

// NewSimpleCommand creates a cobra command from a definition.
func NewSimpleCommand(def SimpleCommandDef) *cobra.Command {
	vals := &flagValues{
		strings:      make(map[string]*string),
		bools:        make(map[string]*bool),
		ints:         make(map[string]*int),
		stringSlices: make(map[string]*[]string),
	}

	cmd := &cobra.Command{
		Use:   def.Use,
		Short: def.Short,
		Long:  def.Long,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			runner := NewRunnerFromCmd(cmd, true)
			params := buildParams(def.Flags, vals)

			result := runner.CallWithParams(def.Method, params)
			successMsg := def.Success
			runner.PrintResult(result, func(any) {
				if successMsg != "" {
					printSuccess(result, successMsg)
				}
			})
		},
	}

	// Add flags based on type
	for _, f := range def.Flags {
		addFlag(cmd, f, vals)
	}

	return cmd
}

func addFlag(cmd *cobra.Command, f Flag, vals *flagValues) {
	switch f.Type {
	case FlagString:
		def := ""
		if f.Default != nil {
			def = f.Default.(string)
		}
		vals.strings[f.Name] = new(string)
		cmd.Flags().StringVarP(vals.strings[f.Name], f.Name, f.Short, def, f.Usage)
	case FlagBool:
		def := false
		if f.Default != nil {
			def = f.Default.(bool)
		}
		vals.bools[f.Name] = new(bool)
		cmd.Flags().BoolVarP(vals.bools[f.Name], f.Name, f.Short, def, f.Usage)
	case FlagInt:
		def := 0
		if f.Default != nil {
			def = f.Default.(int)
		}
		vals.ints[f.Name] = new(int)
		cmd.Flags().IntVarP(vals.ints[f.Name], f.Name, f.Short, def, f.Usage)
	case FlagStringSlice:
		vals.stringSlices[f.Name] = new([]string)
		*vals.stringSlices[f.Name] = []string{}
		cmd.Flags().StringSliceVarP(vals.stringSlices[f.Name], f.Name, f.Short, []string{}, f.Usage)
	}
	if f.Required {
		_ = cmd.MarkFlagRequired(f.Name)
	}
}

func buildParams(flags []Flag, vals *flagValues) map[string]any {
	params := make(map[string]any)
	for _, f := range flags {
		// Use ParamName if set, otherwise use Name
		paramName := f.ParamName
		if paramName == "" {
			paramName = f.Name
		}
		switch f.Type {
		case FlagString:
			if ptr := vals.strings[f.Name]; ptr != nil && *ptr != "" {
				params[paramName] = *ptr
			}
		case FlagBool:
			if ptr := vals.bools[f.Name]; ptr != nil {
				params[paramName] = *ptr
			}
		case FlagInt:
			if ptr := vals.ints[f.Name]; ptr != nil && *ptr != 0 {
				params[paramName] = *ptr
			}
		case FlagStringSlice:
			if ptr := vals.stringSlices[f.Name]; ptr != nil && len(*ptr) > 0 {
				params[paramName] = *ptr
			}
		}
	}
	return params
}

func printSuccess(result any, msg string) {
	if r, ok := result.(map[string]any); ok {
		if success, ok := r["success"].(bool); ok && success {
			fmt.Fprintln(os.Stderr, msg)
		}
	}
}

// Common flag definitions.
var (
	// ToFlag is the standard recipient flag (--to, -t). Maps to "peer" in API.
	ToFlag = Flag{Name: "to", Short: "t", Usage: "Chat/channel (@username or username)", Required: true, ParamName: "peer"}
	// PeerFlag is deprecated, use ToFlag instead. Kept for backwards compatibility.
	PeerFlag        = ToFlag
	UserFlag        = Flag{Name: "user", Short: "u", Usage: "User (@username or username)", Required: true}
	FileFlag        = Flag{Name: "file", Short: "f", Usage: "File path", Required: true}
	TitleFlag       = Flag{Name: "title", Short: "T", Usage: "Title", Required: true}
	MembersFlag = Flag{
		Name: "members", Short: "m", Usage: "Members (can specify multiple)",
		Required: true, Type: FlagStringSlice,
	}
	LinkFlag        = Flag{Name: "link", Short: "l", Usage: "Invite link", Required: true}
	ChannelFlag     = Flag{Name: "channel", Short: "c", Usage: "Channel (@username or username)", Required: true}
	DescriptionFlag = Flag{Name: "description", Short: "d", Usage: "Description"}
	UsernameFlag    = Flag{Name: "username", Short: "U", Usage: "Public username"}
	MegagroupFlag   = Flag{Name: "megagroup", Short: "g", Usage: "Create as supergroup", Type: FlagBool}
)
