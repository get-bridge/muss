package proc

import (
	"io"
	"os"
	"os/exec"
)

// StreamFilter will be called at command delegation to allow filtering output streams.
type StreamFilter interface {
	SetReader(io.Reader)
	SetWriter(io.Writer)
	Start(chan bool)
	Stop()
}

// Delegator holds readers/writers to which to delegate command output.
type Delegator struct {
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
	DoneCh   chan bool
	SignalCh chan os.Signal

	stdoutFilter StreamFilter
	stderrFilter StreamFilter
}

// FilterStdout applies a StreamFilter to stdout.
func (d *Delegator) FilterStdout(f StreamFilter) error {
	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}

	f.SetReader(pr)
	f.SetWriter(d.Stdout)
	d.Stdout = pw
	d.stdoutFilter = f

	return nil
}

// FilterStderr applies a StreamFilter to stderr.
func (d *Delegator) FilterStderr(f StreamFilter) error {
	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}

	f.SetReader(pr)
	f.SetWriter(d.Stderr)
	d.Stderr = pw
	d.stderrFilter = f

	return nil
}

// Delegate runs with a Delegator made from `os.Std*`.
func Delegate(commands ...*exec.Cmd) (err error) {
	return (&Delegator{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}).Delegate(commands...)
}

// Delegate runs the provided commands
// forwarding stdio and signals
// and waits for them to finish.
func (d *Delegator) Delegate(commands ...*exec.Cmd) (err error) {
	// Nothing to do
	if len(commands) == 0 {
		return nil
	}

	d.DoneCh = make(chan bool, 1)

	if d.stdoutFilter != nil {
		d.stdoutFilter.Start(d.DoneCh)
		defer func() {
			if f, ok := d.Stdout.(io.WriteCloser); ok {
				f.Close()
			}
			d.stdoutFilter.Stop()
		}()
	}

	if d.stderrFilter != nil {
		d.stderrFilter.Start(d.DoneCh)
		defer func() {
			if f, ok := d.Stderr.(io.WriteCloser); ok {
				f.Close()
			}
			d.stderrFilter.Stop()
		}()
	}

	cmdch := make(chan error, len(commands))
	for _, cmd := range commands {
		cmd.Stdin = d.Stdin
		cmd.Stdout = d.Stdout
		cmd.Stderr = d.Stderr

		go func(cmd *exec.Cmd) {
			cmdErr := cmd.Run()
			cmdch <- cmdErr
		}(cmd)
	}

	if d.SignalCh == nil {
		d.SignalCh = make(chan os.Signal, 1)
	}
	setupSignals(d.SignalCh)
	defer restoreSignals(d.SignalCh)
	finished := 0

	for {
		select {
		case sig := <-d.SignalCh:
			// Let listening go routines know that we are going to exit.
			if d.DoneCh != nil {
				close(d.DoneCh)
				// Don't try to close it again.
				d.DoneCh = nil
			}

			// Don't forward SIGINT since ctrl-c will go to the process group.
			// Sending again would double the number of signals the child receives.
			if sig == os.Interrupt {
				continue
			}

			for _, cmd := range commands {
				cmd.Process.Signal(sig)
			}

			// Go back to the loop and wait for the commands to finish.
			continue

		case cmdErr := <-cmdch:
			finished++

			// So far we only delegate to one command at time.
			// If we ever do multiple we might want to return a value
			// that can link the error to the command it came from.
			if cmdErr != nil && err == nil {
				err = cmdErr
			}

			if finished == len(commands) {
				return
			}
		}
	}
}
