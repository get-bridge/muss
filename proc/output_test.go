package proc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmdOutput(t *testing.T) {
	t.Run("returns stdout/err", func(t *testing.T) {
		stdout, stderr, err := CmdOutput("/bin/sh", "-c", "echo err >&2; echo out")

		assert.Equal(t, "err", stderr)
		assert.Equal(t, "out", stdout)
		assert.Nil(t, err)
	})

	t.Run("returns err", func(t *testing.T) {
		stdout, stderr, err := CmdOutput("/bin/sh", "-c", "printf 'a\nb\n' >&2; echo out; echo more; echo; exit 1")

		assert.Equal(t, "a\nb", stderr)
		assert.Equal(t, "out\nmore", stdout)
		assert.Equal(t, err.Error(), "exit status 1")
	})
}
