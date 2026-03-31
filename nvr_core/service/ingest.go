package service

import (
	"context"
	"log"
	"time"

	"nvr_core/db/models"
	"nvr_core/db/repository"
)

// IngestService handles the high-throughput buffering of segment metadata.
type IngestService interface {
	Enqueue(seg *models.Segment)
	Start(ctx context.Context)
}

type batchIngester struct {
	repo       repository.SegmentRepository
	bufferChan chan *models.Segment
	batchSize  int
}


// NewBatchIngester initializes the background ingestion service.
func NewBatchIngester(repo repository.SegmentRepository, bufferSize int, batchSize int) IngestService {
	return &batchIngester{
		repo:       repo,
		bufferChan: make(chan *models.Segment, bufferSize),
		batchSize:  batchSize,
	}
}

// Enqueue accepts a new segment from the C++ IPC pipeline.
func (b *batchIngester) Enqueue(seg *models.Segment) {
	b.bufferChan <- seg
}

// Start runs the background flush loop.
func (b *batchIngester) Start(ctx context.Context) {
	var batch []*models.Segment
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			b.flush(batch) // Final flush on shutdown
			return
		case seg := <-b.bufferChan:
			batch = append(batch, seg)
			if len(batch) >= b.batchSize {
				b.flush(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				b.flush(batch)
				batch = nil
			}
		}
	}
}

func (b *batchIngester) flush(batch []*models.Segment) {
	if len(batch) == 0 {
		return
	}

	// Delegate the actual database write to the Repository
	if err := b.repo.BulkInsert(context.Background(), batch); err != nil {
		log.Printf("[IngestService] Failed to flush batch to database: %v", err)
	}
}