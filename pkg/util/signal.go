package util

import (
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalHandler creates a channel that will be closed when termination signals are received
func SetupSignalHandler() <-chan struct{} {
	stopCh := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		close(stopCh)
		<-c
		os.Exit(1)
	}()
	return stopCh
}
