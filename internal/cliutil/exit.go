package cliutil

import "os"

var exitFunc = os.Exit

// SetExitFunc overrides process termination for tests.
func SetExitFunc(fn func(int)) {
	if fn == nil {
		exitFunc = os.Exit
		return
	}
	exitFunc = fn
}

// Exit terminates the process. It is wrapped so CLI helpers can be tested.
func Exit(code int) {
	exitFunc(code)
}
