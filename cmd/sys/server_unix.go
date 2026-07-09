//go:build !windows

package sys

import (
	"os/exec"
	"syscall"
)

func detachServerProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
