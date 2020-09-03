package proc

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/instructure-bridge/muss/testutil"
)

func TestDelegator(t *testing.T) {
	t.Run("func", func(t *testing.T) {
		var err error
		stderr := testutil.CaptureStderr(t, func() {
			err = Delegate(exec.Command("/bin/sh", "-c", "echo dark >&2; exit 1"))
		})

		assert.Equal(t, "dark\n", stderr)
		if err == nil {
			t.Fatal("expected err got nil")
		}
		assert.Equal(t, "exit status 1", err.Error())
	})

	t.Run("bytes buffer", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		d := &Delegator{
			Stdout: &stdout,
			Stderr: &stderr,
		}

		d.Delegate(
			exec.Command("/bin/sh", "-c", "echo err >&2; echo out"),
		)

		assert.Equal(t, "err\n", string(stderr.Bytes()))
		assert.Equal(t, "out\n", string(stdout.Bytes()))
	})

	t.Run("multiple commands", func(t *testing.T) {
		stdout := testutil.TempFile(t, "", "muss-proc-out")
		stderr := testutil.TempFile(t, "", "muss-proc-err")

		defer os.Remove(stdout.Name())
		defer os.Remove(stderr.Name())

		d := &Delegator{
			Stdout: stdout,
			Stderr: stderr,
		}

		d.Delegate(
			exec.Command("/bin/sh", "-c", "echo err >&2; echo out"),
			exec.Command("/bin/sh", "-c", "sleep 1; echo two >&2; echo one"),
		)

		assert.Equal(t, "err\ntwo\n", testutil.ReadFile(t, stderr.Name()))
		assert.Equal(t, "out\none\n", testutil.ReadFile(t, stdout.Name()))
	})

	t.Run("stream filter", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		d := &Delegator{
			Stdout: &stdout,
			Stderr: &stderr,
		}
		fout := newTestFilter()
		ferr := newTestFilter()
		d.FilterStdout(fout)
		d.FilterStderr(ferr)

		d.Delegate(
			exec.Command("/bin/sh", "-c", "echo a >&2; echo A; echo b >&2; echo B"),
		)

		assert.Equal(t, "1 a\n2 b\ndone\n", stderr.String())
		assert.Equal(t, "1 A\n2 B\ndone\n", stdout.String())

		assert.Equal(t, []string{"a", "b"}, ferr.(*testFilter).messages)
		assert.Equal(t, []string{"A", "B"}, fout.(*testFilter).messages)
	})

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

type testFilter struct {
	*Pipe
	messages     []string
	readerDoneCh chan bool
}

func newTestFilter() StreamFilter {
	return &testFilter{
		Pipe:         &Pipe{},
		readerDoneCh: make(chan bool),
	}
}

func (f *testFilter) Stop() {
	<-f.readerDoneCh
	f.Writer().Write([]byte("done\n"))
}

func (f *testFilter) Start(done chan bool) {
	reader := f.Reader()
	writer := f.Writer()

	i := 0
	go func() {
		if f, ok := reader.(io.ReadCloser); ok {
			defer f.Close()
		}
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Bytes()
			i = i + 1
			writer.Write([]byte(fmt.Sprintf("%d %s\n", i, line)))
			f.messages = append(f.messages, string(line))
		}
		f.readerDoneCh <- true
	}()
}
