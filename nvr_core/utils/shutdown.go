package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalContext returns a context that automatically cancels 
// when the application receives an interrupt or termination signal.
func SetupSignalContext() (context.Context, context.CancelFunc) {
	// context.Background() is the empty root context.
	// NotifyContext wraps it, listening for SIGINT and SIGTERM.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	return ctx, cancel
}