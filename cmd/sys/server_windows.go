//go:build windows

package sys

import "os/exec"

func detachServerProcess(_ *exec.Cmd) {}
