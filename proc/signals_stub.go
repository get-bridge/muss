// +build !darwin,!linux

package proc

import (
	"os"
	"os/exec"
)

func sendSignal(proc *os.Process, sig os.Signal) error {
	return proc.Signal(sig)
}

func prepareCommand(cmd *exec.Cmd) {
	// noop
}

func waitForProcessGroup(pgid int) {
	// noop
}
