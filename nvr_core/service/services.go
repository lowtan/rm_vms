package service

import (
	"context"
	"database/sql"
	"nvr_core/db/repository"
)

// Services acts as a dependency injection container for the API layer.
// The API layer knows NOTHING about SQLite or Repositories, only these interfaces.
type Services struct {
	Timeline   TimelineService
	Playback   PlaybackService
	// Camera   service.CameraService
	System     SystemService
}



func NewServices(dbConn *sql.DB) *Services {

	segRepo := repository.NewSegmentRepository(dbConn)
	timelineSvc := NewTimelineService(segRepo)
	playbackSvc := NewPlaybackService(segRepo)

	return &Services{
		Timeline: timelineSvc,
		Playback: playbackSvc,
	}
}

func StartIngester(ctx context.Context, dbConn *sql.DB) IngestService {

	repo := repository.NewSegmentRepository(dbConn)

	// Initialize the Global BatchIngester
	// Buffer 200 segments, flush to disk in batches of 50
	ingester := NewBatchIngester(repo, 200, 50)
	go ingester.Start(ctx)

	return ingester

}
