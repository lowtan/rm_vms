package utils

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// WaitForExitSignal blocks execution until the OS sends a termination signal
// (Ctrl+C, Docker Stop, Kubernetes Terminate, etc.)
func WaitForExitSignal() {

	c := make(chan os.Signal, 1)

	// Subscribe to signals
	// SIGINT  = Ctrl+C
	// SIGTERM = Standard "Kill" command (Docker stop)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block here until a message arrives
	<-c

	fmt.Println("\n[Signal] Shutdown signal received. Cleaning up...")
}