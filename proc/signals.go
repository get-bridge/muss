package proc

import (
	"os"
	"os/signal"
	"syscall"
)

func setupSignals(c chan os.Signal) {
	signals := []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM}
	for _, sig := range signals {
		if !signal.Ignored(sig) {
			signal.Notify(c, sig)
		}
	}
}

func restoreSignals(c chan os.Signal) {
	signal.Stop(c)
}
