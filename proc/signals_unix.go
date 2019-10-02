// +build darwin linux

package proc

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

func sendSignal(proc *os.Process, sig os.Signal) error {
	return syscall.Kill(-proc.Pid, sig.(syscall.Signal))
}

func prepareCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func waitForProcessGroup(pgid int) error {
	var status unix.WaitStatus

	// In manual testing I randomly get either success or "no child process"
	// which is probably from one of the several children it waited for
	// which was probably reaped in the other loop or by its parent.
	// The effect is the same: it actually waits for docker-compose to finish.
	if _, err := unix.Wait4(-pgid, &status, 0, nil); err != nil && err != syscall.ECHILD {
		return err
	}
	return nil
}
