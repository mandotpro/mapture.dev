//go:build windows

package cmd

import (
	"os"
	"syscall"
)

func serveSignals() []os.Signal {
	return []os.Signal{os.Interrupt, syscall.SIGTERM}
}
