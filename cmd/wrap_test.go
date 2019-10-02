package cmd

import (
	"os"
	"sort"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWrapCommand(t *testing.T) {
	withTestPath(t, func(*testing.T) {
		t.Run("multiple commands", func(*testing.T) {
			stdout, stderr, err := testCmdBuilder(newWrapCommand, []string{
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

		t.Run("wait for process groups", func(*testing.T) {

			// Give the cmd 1s to start up then sent interrupt.
			go func() {
				time.Sleep(1 * time.Second)
				syscall.Kill(os.Getpid(), syscall.SIGINT)
			}()

			stdout, stderr, err := testCmdBuilder(newWrapCommand, []string{
				"-s", "/bin/sh",
				"-c", "out () { sleep 1; echo c >&2; }; trap out INT; sleep 15",
				"-c", `$0 -c "out () { sleep 1; echo a; }; trap out TERM; sleep 16 & wait" & pids=$!; $0 -c "out () { sleep 2; echo b; }; trap out TERM; sleep 17 & wait" & pids="$pids $!"; all () { kill -s TERM $pids; wait; }; trap all INT; sleep 18`,
			})

			expOut := "a\nb\n"

			assert.Nil(t, err)
			assert.Equal(t, "c\n", stderr)
			assert.Equal(t, expOut, stdout, "args passed through")
		})

		t.Run("usage errors", func(*testing.T) {

			assert.Contains(t, errFromWrapCmd(t, "-c", "echo", "--exec"),
				"--exec and -c are mutually exclusive")

			assert.Contains(t, errFromWrapCmd(t, "--exec"),
				"--exec requires a command")

		})
	})
}

func errFromWrapCmd(t *testing.T, args ...string) string {
	stdout, stderr, err := testCmdBuilder(newWrapCommand, args)

	assert.NotNil(t, err)

	// Error is on stdout.
	assert.Equal(t, "", stderr, "no stderr")

	return stdout
}
