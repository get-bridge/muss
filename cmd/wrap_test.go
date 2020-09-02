package cmd

import (
	"os"
	"sort"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/instructure-bridge/muss/proc"
)

func TestWrapCommand(t *testing.T) {
	withTestPath(t, func(t *testing.T) {
		t.Run("multiple commands", func(t *testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{
				"wrap",
				"-c", "echo foo",
				"-c", "echo bar",
				"-c", "echo err >&2",
				"echo", "baz",
			})

			// sorted for ease of comparison
			expOut := "bar\nbaz\nfoo\n"
			actual := strings.SplitAfter(stdout, "\n")
			sort.Strings(actual)

			assert.Nil(t, err)
			assert.Equal(t, "err\n", stderr)
			assert.Equal(t, expOut, strings.Join(actual, ""))
		})

		t.Run("wait for processes", func(t *testing.T) {
			// Give the cmd 1s to start up then send signal.
			go func() {
				time.Sleep(1 * time.Second)
				syscall.Kill(os.Getpid(), syscall.SIGTERM)
			}()

			stdout, stderr, err := runTestCommand(nil, []string{
				"wrap",
				"-s", "/bin/sh",
				"-c", "out () { sleep 1; echo c >&2; }; trap out TERM; sleep 5",
				"-c", `$0 -c "out () { sleep 1; echo a; }; trap out TERM; sleep 6 & wait" & pids=$!; $0 -c "out () { sleep 2; echo b; }; trap out TERM; sleep 7 & wait" & pids="$pids $!"; all () { kill -s TERM $pids; wait; }; trap all TERM; sleep 8 & wait; exit 0`,
			})

			expOut := "a\nb\n"

			assert.Nil(t, err)
			assert.Equal(t, "c\n", stderr)
			assert.Equal(t, expOut, stdout, "got output from all")
		})

		t.Run("exec", func(t *testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{
				"wrap",
				"--exec", "echo", "foo",
			})

			assert.Nil(t, err)
			assert.Equal(t, "", stderr)
			assert.Equal(t, "", stdout)

			assert.Equal(t, []string{"echo", "foo"}, proc.LastExecArgv)
		})

		t.Run("shell", func(t *testing.T) {
			shell := os.Getenv("SHELL")
			defer os.Setenv("SHELL", shell)
			os.Unsetenv("SHELL")

			stdout, stderr, err := runTestCommand(nil, []string{
				"wrap",
				"-c", "echo $0",
			})

			assert.Nil(t, err)
			assert.Equal(t, "", stderr)
			assert.Equal(t, "/bin/sh\n", stdout, "defaults to /bin/sh")

			os.Setenv("SHELL", "sh")

			stdout, stderr, err = runTestCommand(nil, []string{
				"wrap",
				"-c", "echo $0",
			})

			assert.Nil(t, err)
			assert.Equal(t, "", stderr)
			assert.Equal(t, "sh\n", stdout, "defaults to $SHELL")
		})

		t.Run("usage errors", func(t *testing.T) {

			assert.Contains(t, errFromWrapCmd(t, "-c", "echo", "--exec"),
				"--exec and -c are mutually exclusive")

			assert.Contains(t, errFromWrapCmd(t, "--exec"),
				"--exec requires a command")

		})
	})
}

func errFromWrapCmd(t *testing.T, args ...string) string {
	stdout, stderr, err := runTestCommand(nil, append([]string{"wrap"}, args...))

	assert.NotNil(t, err)

	// The package Execute() function calls the root command and then prints
	// errors so the subcommand itself won't print anything.
	assert.Equal(t, "", stderr, "no stderr")
	assert.Equal(t, "", stdout, "no stdout")

	return err.Error()
}
