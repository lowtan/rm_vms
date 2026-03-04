package process

import (
	"log"
	"context"
	"nvr_core/utils"
)

const CPP_WORKER_BIN = "./nvr_worker"

func Startup(ctx context.Context, cfg *utils.Config) {

	pm := NewManager(ctx, cfg, 4, CPP_WORKER_BIN)

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

}