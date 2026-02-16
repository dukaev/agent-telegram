package send

import (
	"fmt"
	"os"
)

// resolvePeerAndFile resolves peer and file from positional args.
// With 2 args: args[0]=peer, args[1]=file
// With 1 arg + --to set: args[0]=file
// With 1 arg + --to NOT set: error (file required)
func resolvePeerAndFile(flags *SendFlags, args []string) string {
	switch len(args) {
	case 2:
		_ = flags.To.Set(args[0])
		return args[1]
	case 1:
		if flags.To.Peer() != "" {
			return args[0]
		}
		_ = flags.To.Set(args[0])
		fmt.Fprintln(os.Stderr, "Error: file path is required")
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "Error: peer and file path are required")
	os.Exit(1)
	return ""
}
