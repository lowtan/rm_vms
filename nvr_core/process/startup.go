package process

import (
	"log"
	"context"
	"nvr_core/utils"
	"nvr_core/db"
	"nvr_core/db/ingest"
)

const CPP_WORKER_BIN = "./nvr_worker"

func Startup(ctx context.Context, cfg *utils.Config) (*Manager) {

	ingester := startIngester(ctx)

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

func startIngester(ctx context.Context) *ingest.BatchIngester {

	// Initialize the SQLite Database
	dbConn, err := db.NewConnection("db/nvr_metadata.db")
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}

	// Ensure tables are created
	if err := db.RunMigrations(ctx, dbConn); err != nil {
		log.Fatalf("Failed to run DB migrations: %v", err)
	}

	// Initialize the Global BatchIngester
	// Buffer 200 segments, flush to disk in batches of 50
	ingester := ingest.NewBatchIngester(dbConn, 200, 50)
	go ingester.Start(ctx)

	return ingester

}