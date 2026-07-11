package cliutil

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	unknownLongFlagPattern  = regexp.MustCompile(`^unknown flag: --([^\s]+)$`)
	unknownShortFlagPattern = regexp.MustCompile(`^unknown shorthand flag: .+ in (-[^\s]+)$`)
)

// FlagErrorWithHints enriches unknown-flag parse errors with safe, actionable
// guidance. It never changes argv or retries command execution.
func FlagErrorWithHints(cmd *cobra.Command, err error) error {
	if err == nil {
		return nil
	}

	message := err.Error()
	if match := unknownLongFlagPattern.FindStringSubmatch(message); match != nil {
		var hints []string
		if nearest := nearestLongFlag(cmd, match[1]); nearest != "" {
			hints = append(hints, "did you mean --"+nearest+"?")
		}
		if example := commandExample(cmd); example != "" {
			hints = append(hints, "example:\n"+example)
		}
		return errorWithHints(err, hints)
	}

	if match := unknownShortFlagPattern.FindStringSubmatch(message); match != nil {
		token := match[1]
		if negativePeerPattern.MatchString(token) && !AcceptsFirstArgPeer(cmd) {
			return errorWithHints(err, []string{
				fmt.Sprintf("negative peer IDs must use --to=%s if this command does not accept a positional peer", token),
			})
		}
		return err
	}

	return err
}

func errorWithHints(err error, hints []string) error {
	if len(hints) == 0 {
		return err
	}
	return fmt.Errorf("%w\n\n%s", err, strings.Join(hints, "\n"))
}

func commandExample(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	if example := strings.TrimSpace(cmd.Example); example != "" {
		return example
	}
	if cmd.Name() == "step" && cmd.Parent() != nil && cmd.Parent().Name() == "bot" {
		return "agent-telegram bot step <peer> --send <text>"
	}
	return ""
}

func nearestLongFlag(cmd *cobra.Command, unknown string) string {
	if cmd == nil {
		return ""
	}
	names := make(map[string]struct{})
	collectLongFlagNames(cmd.Flags(), names)
	collectLongFlagNames(cmd.InheritedFlags(), names)

	bestName := ""
	bestDistance := 3
	unique := true
	for name := range names {
		distance := levenshteinDistance(unknown, name)
		switch {
		case distance < bestDistance:
			bestName = name
			bestDistance = distance
			unique = true
		case distance == bestDistance:
			unique = false
		}
	}
	if !unique || bestDistance > 2 {
		return ""
	}
	return bestName
}

func collectLongFlagNames(flags *pflag.FlagSet, names map[string]struct{}) {
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Name != "" {
			names[flag.Name] = struct{}{}
		}
	})
}

func levenshteinDistance(a, b string) int {
	previous := make([]int, len(b)+1)
	for j := range previous {
		previous[j] = j
	}
	for i := 1; i <= len(a); i++ {
		current := make([]int, len(b)+1)
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			current[j] = min(current[j-1]+1, previous[j]+1, previous[j-1]+cost)
		}
		previous = current
	}
	return previous[len(b)]
}
