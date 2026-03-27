package process

import (
	"context"
	"database/sql"
	"log"
	// "nvr_core/db"
	"nvr_core/db/ingest"
	"nvr_core/utils"
)

const CPP_WORKER_BIN = "./nvr_worker"

func Startup(ctx context.Context, cfg *utils.Config, dbConn *sql.DB) (*Manager) {

	ingester := startIngester(ctx, dbConn)

	pm := NewManager(ctx, cfg, 4, CPP_WORKER_BIN, ingester)

	if err := pm.StartAll(); err != nil {
		log.Fatalf("Failed to start workers: %v", err)
	}

	// Distribute Cameras
	for _, cam := range cfg.Cameras {
		if cam.Enabled {
			if err := pm.AssignCamera(cam.ID, cam.URL); err != nil {
				log.Printf("Failed to assign cam %d: %v", cam.ID, err)
			}
		}
	}

	return pm;

}

func startIngester(ctx context.Context, dbConn *sql.DB) *ingest.BatchIngester {

	// Initialize the Global BatchIngester
	// Buffer 200 segments, flush to disk in batches of 50
	ingester := ingest.NewBatchIngester(dbConn, 200, 50)
	go ingester.Start(ctx)

	return ingester

}