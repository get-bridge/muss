// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package proc

import (
	"os"
	"os/exec"
	"syscall"
)

// EnableExec should be called by the main program to setup the actual
// exec call (else a stub is in place for tests).
func EnableExec() {
	execFrd = syscall.Exec
}

// LastExecArgv is only for use by the test suite.
var LastExecArgv []string

var execFrd = execStub

// Exec replaces the current process with the specified command
// looking it up in the path first.
func Exec(args []string) error {
	path, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	return execFrd(path, args, os.Environ())
}

func execStub(argv0 string, argv []string, envv []string) error {
	LastExecArgv = argv
	return nil
}
