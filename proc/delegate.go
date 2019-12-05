package proc

import (
	"io"
	"os"
	"os/exec"
)

// Delegator holds readers/writers to which to delegate command output.
type Delegator struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	DoneCh chan bool
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

	cmdch := make(chan *exec.Cmd, len(commands))
	for _, cmd := range commands {
		cmd.Stdin = d.Stdin
		cmd.Stdout = d.Stdout
		cmd.Stderr = d.Stderr

		go func(cmd *exec.Cmd) {
			cmd.Run()
			cmdch <- cmd
		}(cmd)
	}

	signals := setupSignals()
	finished := 0

	for {
		select {
		case sig := <-signals:
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

		case <-cmdch:
			finished++
			if finished == len(commands) {
				return
			}
		}
	}
}
