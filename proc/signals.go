package proc

import (
	"os"
	"os/signal"
	"syscall"
)

func setupSignals() chan os.Signal {
	c := make(chan os.Signal, 1)

	signals := []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM}
	for _, sig := range signals {
		if !signal.Ignored(sig) {
			signal.Notify(c, sig)
		}
	}

	return c
}
