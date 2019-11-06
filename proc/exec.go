package proc

import (
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

func execStub(argv0 string, argv []string, envv []string) error {
	LastExecArgv = argv
	return nil
}
