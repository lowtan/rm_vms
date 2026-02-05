package main

import (
	"fmt"
	"log"

	// Internal Packages
	"nvr_core/process"
	"nvr_core/utils"
)

const CPP_WORKER_BIN = "./nvr_worker"

// Matches the JSON sent from C++
type WorkerResponse struct {
	Status string `json:"status"`
	CamID  int    `json:"cam"`
}

func main() {

	// Load Configuration
	fmt.Println("[Go Manager] v.0.0.1")
	fmt.Println("[Go Manager] Loading config...")

	cfg, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	fmt.Printf("[Go Manager] Config Loaded. Storage: %s, Cameras: %d\n", 
			cfg.Server.StoragePath, len(cfg.Cameras))

	pm := process.NewManager(4, CPP_WORKER_BIN)

	if err := pm.StartAll(); err != nil {
		log.Fatalf("Failed to start workers: %v", err)
	}
	defer pm.StopAll() // Cleanup on exit

	// Distribute Cameras
	for _, cam := range cfg.Cameras {
		if cam.Enabled {
			if err := pm.AssignCamera(cam.ID, cam.URL); err != nil {
				log.Printf("Failed to assign cam %d: %v", cam.ID, err)
			}
		}
	}

	// Block until Ctrl+C (Graceful Shutdown)
	utils.WaitForExitSignal()

}