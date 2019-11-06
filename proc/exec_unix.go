// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package proc

import (
	"os"
	"os/exec"
)

// Exec replaces the current process with the specified command
// looking it up in the path first.
func Exec(args []string) error {
	path, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	return execFrd(path, args, os.Environ())
}
