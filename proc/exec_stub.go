// +build !aix,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package proc

import (
	"os/exec"
)

// Exec uses Delegate on systems without `execve`.
func Exec(args []string) error {
	return Delegate(exec.Command(args[0], args[1:]...))
}
