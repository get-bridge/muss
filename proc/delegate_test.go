package proc

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDelegator(t *testing.T) {
	t.Run("signals", func(t *testing.T) {
		var stdout bytes.Buffer
		d := Delegator{
			Stdout: &stdout,
		}
		d.SignalCh = make(chan os.Signal, 1)
		script := `trap 'echo TERM; kill %1; exit 1' TERM; trap 'echo INT' INT; sleep 3 & wait; echo done`

		go func() {
			time.Sleep(1 * time.Second)
			d.SignalCh <- syscall.SIGTERM
		}()

		err := d.Delegate(exec.Command("/bin/sh", "-c", script))

		assert.Equal(t, "TERM\n", stdout.String(), "TERM forwards")
		assert.Equal(t, "exit status 1", err.Error())

		stdout.Reset()

		go func() {
			time.Sleep(1 * time.Second)
			d.SignalCh <- syscall.SIGINT
		}()

		err = d.Delegate(exec.Command("/bin/sh", "-c", script))

		assert.Equal(t, "done\n", stdout.String(), "INT waits for end")
		assert.Nil(t, err)
	})
}
