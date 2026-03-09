package main

import (
	"fmt"
	"log"
	"time"

	// Internal Packages
	"nvr_core/apiserver"
	"nvr_core/process"
	"nvr_core/utils"
)

func main() {

	// application-wide context
	ctx, cancel := utils.SetupSignalContext()
	defer cancel() // Ensures resources are freed when main exits

	// Load Configuration
	fmt.Println("================================================")
	fmt.Println("[Go Manager] v.0.0.1")
	fmt.Println("[Go Manager] Loading config...")

	cfg, err := utils.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	fmt.Printf("[Go Manager] Config Loaded. Storage: %s, Cameras: %d\n", 
			cfg.Server.StoragePath, len(cfg.Cameras))

	pm := process.Startup(ctx, cfg)

	go apiserver.Initiate(ctx, cfg, pm)

	// Block until the context is canceled (SIGINT/SIGTERM received)
	<-ctx.Done()

	fmt.Println("\n[Signal] Shutdown signal received. Gracefully terminating subsystems...")

	// TODO:
	// At this point, the context is canceled. 
	// The C++ workers are automatically receiving kill signals.
	// We can add a brief time.Sleep() here or use a sync.WaitGroup 
	// to give everything a second to flush buffers before main() finally exits.

	time.Sleep(5*time.Second)

}