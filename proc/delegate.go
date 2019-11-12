package proc

import (
	"io"
	"os"
	"os/exec"
	"sync"
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
		prepareCommand(cmd)

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
			// Should we revert to default so they can hit ctrl-c again?

			// Let listening go routines know that we are going to exit.
			if d.DoneCh != nil {
				close(d.DoneCh)
			}

			var wg sync.WaitGroup

			for _, cmd := range commands {
				wg.Add(2) // one for the actual cmd object and one for the process group

				go func(pid int) {
					waitForProcessGroup(pid)
					wg.Done()
				}(cmd.Process.Pid)

				go func(cmd *exec.Cmd) {
					sendSignal(cmd.Process, sig)
					cmd.Wait()
					wg.Done()
				}(cmd)
			}

			wg.Wait()
			return
		case <-cmdch:
			finished++
			if finished == len(commands) {
				return
			}
		}
	}
}
